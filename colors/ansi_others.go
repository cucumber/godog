// Copyright 2014 shiena Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build !windows
// +build !windows

package colors

import "io"

type ansiColorWriter struct {
	w    io.WriteCloser
	mode outputMode
}

func (cw *ansiColorWriter) Write(p []byte) (int, error) {
	return cw.w.Write(p)
}
func (cw *ansiColorWriter) Close() error {
	return cw.w.Close()
}
