package tidy

type Option func(*Options)

type Options struct {
	SourcePath             string
	TargetPath             string
	Pattern                string
	DryRun                 bool
	DeleteSrc              bool
	DuplicationAutoReplace bool
}

func newOptions(options ...Option) Options {
	opt := Options{}

	for _, o := range options {
		o(&opt)
	}

	return opt
}

func SourcePath(src string) Option {
	return func(o *Options) {
		o.SourcePath = src
	}
}

func TargetPath(dst string) Option {
	return func(o *Options) {
		o.TargetPath = dst
	}
}

func Pattern(pattern string) Option {
	return func(o *Options) {
		o.Pattern = pattern
	}
}

func DryRun(d bool) Option {
	return func(o *Options) {
		o.DryRun = d
	}
}

func DeleteSrc(d bool) Option {
	return func(o *Options) {
		o.DeleteSrc = d
	}
}

func DuplicationAutoReplace(d bool) Option {
	return func(o *Options) {
		o.DuplicationAutoReplace = d
	}
}
