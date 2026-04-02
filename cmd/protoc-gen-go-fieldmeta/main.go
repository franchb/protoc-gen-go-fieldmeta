package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/franchb/protoc-gen-go-fieldmeta/internal/generator"
	"github.com/franchb/protoc-gen-go-fieldmeta/internal/version"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

func main() {
	showVersion := flag.Bool("version", false, "print the version and exit")
	var flags flag.FlagSet

	protogen.Options{ParamFunc: flags.Set}.Run(func(gen *protogen.Plugin) error {
		if *showVersion {
			fmt.Fprintf(os.Stderr, "protoc-gen-go-fieldmeta %s\n", version.Version)
			os.Exit(0)
		}

		gen.SupportedFeatures = uint64(
			pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL |
				pluginpb.CodeGeneratorResponse_FEATURE_SUPPORTS_EDITIONS,
		)
		gen.SupportedEditionsMinimum = descriptorpb.Edition_EDITION_PROTO2
		gen.SupportedEditionsMaximum = descriptorpb.Edition_EDITION_2023

		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			if err := generator.GenerateFile(gen, f); err != nil {
				return err
			}
		}
		return nil
	})
}
