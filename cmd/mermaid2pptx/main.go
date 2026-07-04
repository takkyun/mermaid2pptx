package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"mermaid2pptx/internal/convert"
)

func main() {
	var (
		out    string
		font   string
		force  bool
		margin float64
		mmdc   string
	)
	flag.StringVar(&out, "o", "", "output .pptx path (only with a single input; default: <input>.pptx)")
	flag.StringVar(&font, "font", "Noto Sans JP", "font family used for all text")
	flag.BoolVar(&force, "f", false, "overwrite the output file if it exists")
	flag.Float64Var(&margin, "margin", 0.3, "slide margin in inches")
	flag.StringVar(&mmdc, "mmdc", "", "path to the mermaid-cli binary used for .mmd inputs (default: mmdc in PATH)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] input.{svg|mmd} [input2 ...]\n\nConverts mermaid diagrams (SVG, or .mmd via mermaid-cli) into editable PowerPoint slides.\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	inputs := flag.Args()
	if len(inputs) == 0 {
		flag.Usage()
		os.Exit(2)
	}
	if out != "" && len(inputs) > 1 {
		fmt.Fprintln(os.Stderr, "error: -o cannot be used with multiple inputs")
		os.Exit(2)
	}

	for _, in := range inputs {
		dst := out
		if dst == "" {
			dst = strings.TrimSuffix(in, filepath.Ext(in)) + ".pptx"
		}
		if err := convertFile(in, dst, font, margin, force, mmdc); err != nil {
			fmt.Fprintf(os.Stderr, "error: %s: %v\n", in, err)
			os.Exit(1)
		}
		fmt.Printf("%s -> %s\n", in, dst)
	}
}

func convertFile(in, dst, font string, margin float64, force bool, mmdc string) error {
	if !force {
		if _, err := os.Stat(dst); err == nil {
			return fmt.Errorf("output %s already exists (use -f to overwrite)", dst)
		}
	}
	svgPath := in
	switch strings.ToLower(filepath.Ext(in)) {
	case ".mmd", ".mermaid":
		rendered, err := renderMermaid(in, mmdc)
		if err != nil {
			return err
		}
		defer os.Remove(rendered)
		svgPath = rendered
	}
	f, err := os.Open(svgPath)
	if err != nil {
		return err
	}
	defer f.Close()
	d, err := convert.ParseMermaidSVG(f)
	if err != nil {
		return err
	}
	slideXML := convert.GenerateSlideXML(d, convert.Options{Font: font, MarginIn: margin})
	o, err := os.Create(dst)
	if err != nil {
		return err
	}
	if err := convert.WritePPTX(o, slideXML); err != nil {
		o.Close()
		os.Remove(dst)
		return err
	}
	return o.Close()
}

// renderMermaid renders a .mmd file to a temporary SVG using mermaid-cli.
// The caller removes the returned file.
func renderMermaid(in, mmdc string) (string, error) {
	if mmdc == "" {
		p, err := exec.LookPath("mmdc")
		if err != nil {
			return "", fmt.Errorf(".mmd input requires mermaid-cli; install it with `npm install -g @mermaid-js/mermaid-cli`, pass -mmdc <path>, or convert to SVG yourself")
		}
		mmdc = p
	}
	tmp, err := os.CreateTemp("", "mermaid2pptx-*.svg")
	if err != nil {
		return "", err
	}
	tmp.Close()
	cmd := exec.Command(mmdc, "-i", in, "-o", tmp.Name(), "-I", "my-svg", "-b", "white")
	if outBytes, err := cmd.CombinedOutput(); err != nil {
		os.Remove(tmp.Name())
		return "", fmt.Errorf("mmdc failed: %v\n%s", err, outBytes)
	}
	return tmp.Name(), nil
}
