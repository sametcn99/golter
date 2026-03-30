package converter

import "context"

// Options holds parameters for the conversion.
type Options struct {
	Quality    string
	PandocArgs []string
	EbookArgs  []string
}

// Converter is the interface that all converters must implement.
type Converter interface {
	// Name returns the display name of the converter.
	Name() string
	// CanConvert checks if the converter can handle the conversion from srcExt to targetExt.
	CanConvert(srcExt, targetExt string) bool
	// SupportedSourceExtensions returns a list of source extensions supported by the converter.
	SupportedSourceExtensions() []string
	// SupportedTargetFormats returns a list of target extensions supported for the given source extension.
	SupportedTargetFormats(srcExt string) []string
	// Convert performs the conversion.
	Convert(ctx context.Context, src, target string, opts Options) error
}
