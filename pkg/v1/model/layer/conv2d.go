package layer

import (
	"fmt"

	g "gorgonia.org/gorgonia"
	t "gorgonia.org/tensor"
)

// Conv2D is a 2D convolution.
type Conv2D struct {
	// Input size.
	// required
	Input int

	// Output size.
	// required
	Output int

	// Height of the filter.
	// required
	Height int

	// Width of the filter.
	// required
	Width int

	// Name of the layer.
	Name string

	// Activation function for the layer.
	// Defaults to ReLU
	Activation ActivationFn

	// KernelShape is the shape of the kernec.
	// Defaults to (3, 3)
	KernelShape t.Shape

	// Pad
	// Defaults to (1, 1)
	Pad []int

	// Stride
	// Defaults to (1, 1)
	Stride []int

	// Dilation
	// Defaults to (1, 1)
	Dilation []int

	// Init function fot the weights.
	// Defaults to GlorotN(1)
	Init g.InitWFn
}

// Compile the config into a layer.
func (c Conv2D) Compile(graph *g.ExprGraph, opts ...CompileOpt) Layer {
	cnv := newConv2D(&c)
	for _, opt := range opts {
		opt(cnv)
	}
	if cnv.shared != nil {
		cnv.filter = g.NewTensor(graph, cnv.dtype, 4, g.WithShape(cnv.filterShape...), g.WithInit(c.Init), g.WithName(c.Name), g.WithValue(cnv.shared.filter.Value()))
		return cnv
	}
	cnv.filter = g.NewTensor(graph, cnv.dtype, 4, g.WithShape(cnv.filterShape...), g.WithInit(c.Init), g.WithName(c.Name))
	return cnv
}

// Validate the config.
func (c Conv2D) Validate() error {
	if c.Input == 0 {
		return fmt.Errorf("input must be set")
	}
	if c.Output == 0 {
		return fmt.Errorf("output must be set")
	}
	if c.Width == 0 {
		return fmt.Errorf("width must be set")
	}
	if c.Height == 0 {
		return fmt.Errorf("height must be set")
	}
	return nil
}

// ApplyDefaults to the config.
func (c Conv2D) ApplyDefaults() Config {
	if c.Activation == nil {
		c.Activation = ReLU
	}
	if len(c.KernelShape) == 0 {
		c.KernelShape = []int{3, 3}
	}
	if len(c.Pad) == 0 {
		c.Pad = []int{1, 1}
	}
	if len(c.Stride) == 0 {
		c.Stride = []int{1, 1}
	}
	if len(c.Dilation) == 0 {
		c.Dilation = []int{1, 1}
	}
	if c.Init == nil {
		c.Init = g.GlorotN(1)
	}
	return c
}

// Clone the config.
func (c Conv2D) Clone() Config {
	return Conv2D{
		Input:       c.Input,
		Output:      c.Output,
		Height:      c.Height,
		Width:       c.Width,
		Name:        c.Name,
		Activation:  c.Activation.Clone(),
		KernelShape: c.KernelShape,
		Pad:         c.Pad,
		Stride:      c.Stride,
		Dilation:    c.Dilation,
		Init:        c.Init,
	}
}

// conv2D is a two dimensional convolution layer.
type conv2D struct {
	*Conv2D

	dtype       t.Dtype
	filterShape t.Shape
	filter      *g.Node
	shared      *conv2D
	isBatched   bool
}

func newConv2D(config *Conv2D) *conv2D {
	config.ApplyDefaults()
	return &conv2D{
		Conv2D:      config,
		dtype:       t.Float32,
		filterShape: []int{config.Output, config.Input, config.Height, config.Width},
	}
}

// Fwd is a forward pass through the layer.
func (c *conv2D) Fwd(x *g.Node) (*g.Node, error) {
	n, err := g.Conv2d(x, c.filter, c.KernelShape, c.Pad, c.Stride, c.Dilation)
	if err != nil {
		return nil, err
	}
	return c.Activation.Fwd(n)
}

// Learnables returns all learnable nodes within this layer.
func (c *conv2D) Learnables() g.Nodes {
	return g.Nodes{c.filter}
}

// Clone the layer.
func (c *conv2D) Clone() Layer {
	return &conv2D{
		Conv2D:    c.Conv2D.Clone().(*Conv2D),
		dtype:     c.dtype,
		filter:    c.filter,
		isBatched: c.isBatched,
	}
}

// Graph returns the graph for this layer.
func (c *conv2D) Graph() *g.ExprGraph {
	if c.filter == nil {
		return nil
	}
	return c.filter.Graph()
}
