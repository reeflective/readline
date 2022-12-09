# go-caret

Package **caret** encodes, and decodes caret text (i.e., where ASCII control codes are in caret notation) into UTF-8 text (which may also be ASCII text).

## Documention

Online documentation, which includes examples, can be found at: http://godoc.org/github.com/reiver/go-caret

[![GoDoc](https://godoc.org/github.com/reiver/go-caret?status.svg)](https://godoc.org/github.com/reiver/go-caret)

## ANSI Escape Codes

This can be useful for writing, or reading text that contains ANSI escape codes, such as ANSI color codes.

Note that Caret Notation code note have a good way of representing the caret in the output text.

This is a limit in inherent to Caret Notation.

## Caret Notation

The mapping from caret notation to control codes is as follows:

	^@ ⇒ NUL (0x00)

	^A ⇒ SOH (0x01)

	^B ⇒ STX (0x02)

	^C ⇒ ETX (0x03)

	^D ⇒ EOT (0x04)

	^E ⇒ ENQ (0x05)

	^F ⇒ ACK (0x06)

	^G ⇒ BEL (0x07)

	^H ⇒ BS  (0x08)

	^I ⇒ HT  (0x09)

	^J ⇒ LF  (0x0a)

	^K ⇒ VT  (0x0b)

	^L ⇒ FF  (0x0c)

	^M ⇒ CR  (0x0d)

	^N ⇒ SO  (0x0e)

	^O ⇒ SI  (0x0f)

	^P ⇒ DLE (0x10)

	^Q ⇒ DC1 (0x11)

	^R ⇒ DC2 (0x12)

	^S ⇒ DC3 (0x13)

	^T ⇒ DC4 (0x14)

	^U ⇒ NAK (0x15)

	^V ⇒ SYN (0x16)

	^W ⇒ ETB (0x17)

	^X ⇒ CAN (0x18)

	^Y ⇒ EM  (0x19)

	^Z ⇒ SUB (0x1a)

	^[ ⇒ ESC (0x1b)

	^\ ⇒ FS  (0x1c)

	^] ⇒ GS  (0x1d)

	^^ ⇒ RS	 (0x1e)

	^_ ⇒ US	 (0x1f)

	^` ⇒ SP  (0x20)

	^? ⇒ DEL (0x7f)
