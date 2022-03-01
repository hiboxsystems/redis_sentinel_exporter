# Ensure that 'all' is the default target otherwise it will be the first target from Makefile.common.
all::

include Makefile.common

.PHONY: crossbuild_tarballs release

crossbuild_tarballs:
	promu crossbuild tarballs

release:
	promu release
