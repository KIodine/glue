package glue

// The options control how `Glue` behaves.
type glueOptions struct {
	FavorSource bool
	Strict      bool
}

// The interface all option must implement.
type GlueOption interface {
	apply(*glueOptions)
}

type optFavorSource struct{}

// singleton
var optFavor = &optFavorSource{}

// `Glue` "pushes" field of source to destination struct instead.
func DoFavorSource() GlueOption {
	return optFavor
}

func (*optFavorSource) apply(opt *glueOptions) {
	opt.FavorSource = true
}

type optStrict struct{}

// singleton
var optStrct = &optStrict{}

// `Glue` will return error is one single field is not satisfied.
func DoStrict() GlueOption {
	return optStrct
}

func (*optStrict) apply(opt *glueOptions) {
	opt.Strict = true
}
