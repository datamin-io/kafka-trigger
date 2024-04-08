package command

func getBaseUrl(env string) string {
	switch env {
	case "production":
		return "https://api.datamin.io"
	case "test":
		return "https://api-test.datamin.io"
	}

	panic("unknown environment specified")
}
