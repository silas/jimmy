package migrations

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

type AddProtoInput struct {
	ID   int
	Name string
	Path string
}

func (ms *Migrations) AddProto(_ context.Context, input AddProtoInput) error {
	m, err := ms.Get(input.ID)
	if err != nil {
		return err
	}

	err = checkFile(input.Path, "proto")
	if err != nil {
		return err
	}

	b, err := os.ReadFile(input.Path)
	if err != nil {
		return err
	}

	fileDescriptorSet := &descriptorpb.FileDescriptorSet{}

	err = proto.Unmarshal(b, fileDescriptorSet)
	if err != nil {
		return fmt.Errorf("failed to unmarshal %q file descriptor set: %w", input.Path, err)
	}

	if m.data.FileDescriptorSets == nil {
		m.data.FileDescriptorSets = map[string]*descriptorpb.FileDescriptorSet{}
	}

	for _, file := range fileDescriptorSet.File {
		file.SourceCodeInfo = nil
	}

	m.data.FileDescriptorSets[input.Name] = fileDescriptorSet

	err = Marshal(m.Path(), m.data)
	if err != nil {
		return err
	}

	return nil
}
