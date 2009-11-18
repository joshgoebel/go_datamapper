# Copyright 2009 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

include $(GOROOT)/src/Make.$(GOARCH)

TARG=dm
GOFILES=\
	dm.go\

include $(GOROOT)/src/Make.pkg

states: dm.a
	$(GC) states.go
	$(LD) -o $@ states.$O
	
dm.a: dm.go
	$(GC) dm.go
