package apis

type Chart struct {
	Registry           string
	Repository         string
	Version            string
	AppVersion         string
	AppVersionPrevious string
}

type Repoes struct {
	ImageName string
	Chart
}
