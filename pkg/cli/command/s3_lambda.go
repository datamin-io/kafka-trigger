package command

import (
	"context"
	"fmt"
	"integration/config"
	"integration/pkg/workflows"
	"io"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gobwas/glob"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var runS3ListenerLambdaHandler cli.ActionFunc = func(c *cli.Context) error {
	cfg := config.Cfg()
	apiCfg := cfg.API
	s3Cfg := cfg.S3

	apiClient := workflows.NewClient(getBaseUrl(cfg.Env), apiCfg.ClientId, apiCfg.ClientSecret, apiCfg.BasicAuthUsername, apiCfg.BasicAuthPassword)
	mapping, err := parseS3Mapping(s3Cfg.Mapping)
	if err != nil {
		return err
	}

	lambda.StartWithOptions(getS3EventHandler(apiClient, mapping), lambda.WithContext(c.Context))

	return nil
}

func getS3EventHandler(apiClient workflows.Client, mapping map[string]mappingItem) func(ctx context.Context, e *events.S3Event) (*string, error) {
	return func(ctx context.Context, e *events.S3Event) (*string, error) {
		for _, r := range e.Records {
			fullPath := fmt.Sprintf("%s/%s", r.S3.Bucket.Name, r.S3.Object.Key)
			var content io.Reader
			for _, mi := range mapping {
				pattern := mi.glob
				if !pattern.Match(fullPath) {
					continue
				}

				if content == nil {
					sess, err := session.NewSession(&aws.Config{
						Region: aws.String(r.AWSRegion)},
					)

					if err != nil {
						return nil, err
					}

					s3Client := s3.New(sess)
					requestInput := &s3.GetObjectInput{
						Bucket: aws.String(r.S3.Bucket.Name),
						Key:    aws.String(r.S3.Object.Key),
					}

					result, err := s3Client.GetObject(requestInput)
					if err != nil {
						return nil, err
					}

					defer result.Body.Close()

					content = result.Body
				}

				for _, wfUuid := range mi.uuids {
					runUuid, err := apiClient.RunWorkflow(wfUuid.String(), content)
					if err != nil {
						log.Error(err)
						return nil, nil
					}

					log.Infof("Running workflow %s, run UUID %s", wfUuid.String(), runUuid)
				}
			}
		}

		return nil, nil
	}
}

func parseS3Mapping(raw string) (map[string]mappingItem, error) {
	mapping := map[string]mappingItem{}
	parts := strings.Split(raw, ";")
	for _, pt := range parts {
		parts2 := strings.Split(pt, ":")
		if len(parts2) != 2 {
			return nil, fmt.Errorf("wrong mapping format: %s", pt)
		}

		pattern := parts2[0]
		wfIds := strings.Split(parts2[1], ",")
		if len(wfIds) == 0 {
			return nil, fmt.Errorf("no workflows specified: %s", pt)
		}

		if _, ok := mapping[pattern]; !ok {
			mapping[pattern] = mappingItem{
				glob:  glob.MustCompile(pattern, '/'),
				uuids: []uuid.UUID{},
			}

		}

		for _, wfIdStr := range wfIds {
			mi := mapping[pattern]
			mi.uuids = append(mi.uuids, uuid.MustParse(wfIdStr))
			mapping[pattern] = mi
		}
	}

	return mapping, nil
}

type mappingItem struct {
	glob  glob.Glob
	uuids []uuid.UUID
}

var RunS3ListenerLambdaCommand = &cli.Command{
	Name:        "run-s3-listener-lambda",
	Description: "Run Datamin S3 listener lambda",
	Usage:       "Run Datamin S3 listener lambda",
	Action:      runS3ListenerLambdaHandler,
}
