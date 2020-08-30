package utils

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCutHtml(t *testing.T) {
	req := require.New(t)

	in := "test test test"
	out, cut := CutHtml(in, 1, 20)
	req.False(cut)
	req.Equal(in, out)

	out, cut = CutHtml(in, 1, 10)
	req.True(cut)
	req.Equal("test test…", out)

	in = "<p>test <br>test test </p>"
	out, cut = CutHtml(in, 1, 10)
	req.True(cut)
	req.Equal("<p>test…</p>", out)

	out, cut = CutHtml(in, 2, 10)
	req.Equal(in, out)
	req.False(cut)

	in = "<p><b><i>test <br>test test</i></b></p>"
	out, cut = CutHtml(in, 2, 5)
	req.True(cut)
	req.Equal("<p><b><i>test <br>test…</i></b></p>", out)

	in = "<p>test <br>test</p><p>test test test test</p> <p>test</p>"
	out, cut = CutHtml(in, 4, 7)
	req.True(cut)
	req.Equal("<p>test <br>test</p><p>test test test…</p>", out)

	in = "<p>проверяем кириллические буквы</p>"
	out, cut = CutHtml(in, 1, 10)
	req.True(cut)
	req.Equal("<p>проверяем…</p>", out)

	in = "<h1>test test test</h1><br>"
	out, cut = CutHtml(in, 1, 40)
	req.True(cut)
	req.Equal("<h1>test…</h1>", out)
}
