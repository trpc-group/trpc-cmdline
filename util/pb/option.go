package pb

type options struct {
	secvEnabled     bool
	pb2ImportPath   map[string]string
	pkg2ImportPath  map[string]string
	descriptorSetIn string
}

// Option is used to store the content of the relevant options.
type Option func(*options)

// WithSecvEnabled enables validation and generates stub code using protoc-gen-secv.
func WithSecvEnabled(enabled bool) Option {
	return func(o *options) {
		o.secvEnabled = enabled
	}
}

// WithPb2ImportPath adds mapping between pb file and import path.
func WithPb2ImportPath(m map[string]string) Option {
	return func(o *options) {
		o.pb2ImportPath = m
	}
}

// WithPkg2ImportPath adds the mapping between package name and import path.
func WithPkg2ImportPath(m map[string]string) Option {
	return func(o *options) {
		o.pkg2ImportPath = m
	}
}

// WithDescriptorSetIn adds the descriptor_set_in option to the command.
func WithDescriptorSetIn(descriptorSetIn string) Option {
	return func(o *options) {
		o.descriptorSetIn = descriptorSetIn
	}
}
