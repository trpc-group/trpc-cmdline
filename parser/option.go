package parser

type options struct {
	aliasOn              bool
	aliasAsClientRPCName bool
	language             string
	rpcOnly              bool
	multiVersion         bool
}

// Option parse option
type Option func(*options)

// WithAliasOn enable alias
func WithAliasOn(enabled bool) Option {
	return func(opts *options) {
		if opts != nil {
			opts.aliasOn = enabled
		}
	}
}

// WithAliasAsClientRPCName sets alias as client rpc name.
func WithAliasAsClientRPCName(enabled bool) Option {
	return func(opts *options) {
		if opts != nil {
			opts.aliasAsClientRPCName = enabled
		}
	}
}

// WithLanguage specify language for further checking
func WithLanguage(lang string) Option {
	return func(opts *options) {
		if opts != nil {
			opts.language = lang
		}
	}
}

// WithRPCOnly enable RPC only
func WithRPCOnly(enabled bool) Option {
	return func(opts *options) {
		if opts != nil {
			opts.rpcOnly = enabled
		}
	}
}

// WithMultiVersion enable multi-version support.
func WithMultiVersion(enabled bool) Option {
	return func(opts *options) {
		if opts != nil {
			opts.multiVersion = enabled
		}
	}

}
