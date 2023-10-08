package descriptor

import "github.com/jhump/protoreflect/desc"

// ProtoFileDescriptor implements the Desc interface.
// Describes all information about a protobuf file.
type ProtoFileDescriptor struct {
	FD *desc.FileDescriptor
}

// GetName implements the Desc interface.
func (p *ProtoFileDescriptor) GetName() string {
	return p.FD.GetName()
}

// GetFullyQualifiedName implements the Desc interface.
func (p *ProtoFileDescriptor) GetFullyQualifiedName() string {
	return p.FD.GetFullyQualifiedName()
}

// GetPackage implements the Desc interface.
func (p *ProtoFileDescriptor) GetPackage() string {
	return p.FD.GetPackage()
}

// GetFileOptions implements the Desc interface.
func (p *ProtoFileDescriptor) GetFileOptions() FileOpt {
	return p.FD.GetFileOptions()
}

// GetDependencies implements the Desc interface.
func (p *ProtoFileDescriptor) GetDependencies() []Desc {
	var descs []Desc
	for _, fd := range p.FD.GetDependencies() {
		descs = append(descs, &ProtoFileDescriptor{FD: fd})
	}
	return descs
}

// GetServices implements the Desc interface.
func (p *ProtoFileDescriptor) GetServices() []ServiceDesc {
	var descs []ServiceDesc
	for _, sd := range p.FD.GetServices() {
		descs = append(descs, &ProtoServiceDescriptor{SD: sd})
	}
	return descs
}

// GetMessageTypes implements the Desc interface.
func (p *ProtoFileDescriptor) GetMessageTypes() []MessageDesc {
	var descs []MessageDesc
	for _, md := range p.FD.GetMessageTypes() {
		descs = append(descs, &ProtoMessageDescriptor{MD: md})
	}
	return descs
}

// ProtoServiceDescriptor implements the ServiceDesc interface.
// Describes all information of an RPC service.
type ProtoServiceDescriptor struct {
	SD *desc.ServiceDescriptor
}

// GetName implements the ServiceDesc interface.
func (p *ProtoServiceDescriptor) GetName() string {
	return p.SD.GetName()
}

// GetMethods implements the ServiceDesc interface.
func (p *ProtoServiceDescriptor) GetMethods() []MethodDesc {
	var descs []MethodDesc
	for _, md := range p.SD.GetMethods() {
		descs = append(descs, &ProtoMethodDescriptor{MD: md})
	}
	return descs
}

// ProtoMethodDescriptor implements the MethodDesc interface.
type ProtoMethodDescriptor struct {
	MD *desc.MethodDescriptor
}

// GetName implements the MethodDesc interface.
func (p *ProtoMethodDescriptor) GetName() string {
	return p.MD.GetName()
}

// GetInputType implements the MethodDesc interface.
func (p *ProtoMethodDescriptor) GetInputType() MessageDesc {
	return &ProtoMessageDescriptor{MD: p.MD.GetInputType()}
}

// GetOutputType implements the MethodDesc interface.
func (p *ProtoMethodDescriptor) GetOutputType() MessageDesc {
	return &ProtoMessageDescriptor{MD: p.MD.GetOutputType()}
}

// IsClientStreaming implements the MethodDesc interface.
func (p *ProtoMethodDescriptor) IsClientStreaming() bool {
	return p.MD.IsClientStreaming()
}

// IsServerStreaming implements the MethodDesc interface.
func (p *ProtoMethodDescriptor) IsServerStreaming() bool {
	return p.MD.IsServerStreaming()
}

// GetSourceInfo implements the MethodDesc interface.
func (p *ProtoMethodDescriptor) GetSourceInfo() SourceInfo {
	return p.MD.GetSourceInfo()
}

// ProtoMessageDescriptor implements the MessageDesc interface.
type ProtoMessageDescriptor struct {
	MD *desc.MessageDescriptor
}

// GetFile implements the MessageDesc interface.
func (p *ProtoMessageDescriptor) GetFile() Desc {
	return &ProtoFileDescriptor{FD: p.MD.GetFile()}
}

// GetFullyQualifiedName implements the MessageDesc interface.
func (p *ProtoMessageDescriptor) GetFullyQualifiedName() string {
	return p.MD.GetFullyQualifiedName()
}
