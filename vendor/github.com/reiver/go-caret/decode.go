package caret

// decode returns the ASCII control characer for the Caret Notation code.
//
// So, for example, this is what would be returned for the following inputs:
//
//	@ ⇒ NUL (0x00)
//
//	A ⇒ SOH (0x01)
//
//	B ⇒ STX (0x02)
//
//	C ⇒ ETX (0x03)
//
//	D ⇒ EOT (0x04)
//
//	E ⇒ ENQ (0x05)
//
//	F ⇒ ACK (0x06)
//
//	G ⇒ BEL (0x07)
//
//	H ⇒ BS  (0x08)
//
//	I ⇒ HT  (0x09)
//
//	J ⇒ LF  (0x0a)
//
//	K ⇒ VT  (0x0b)
//
//	L ⇒ FF  (0x0c)
//
//	M ⇒ CR  (0x0d)
//
//	N ⇒ SO  (0x0e)
//
//	O ⇒ SI  (0x0f)
//
//	P ⇒ DLE (0x10)
//
//	Q ⇒ DC1 (0x11)
//
//	R ⇒ DC2 (0x12)
//
//	S ⇒ DC3 (0x13)
//
//	T ⇒ DC4 (0x14)
//
//	U ⇒ NAK (0x15)
//
//	V ⇒ SYN (0x16)
//
//	W ⇒ ETB (0x17)
//
//	X ⇒ CAN (0x18)
//
//	Y ⇒ EM  (0x19)
//
//	Z ⇒ SUB (0x1a)
//
//	[ ⇒ ESC (0x1b)
//
//	\ ⇒ FS  (0x1c)
//
//	] ⇒ GS  (0x1d)
//
//	^ ⇒ RS	(0x1e)
//
//	_ ⇒ US	(0x1f)
//
//	` ⇒ SP	(0x20)
//
//	? ⇒ DEL (0x7f)
//
// Each of these map to the value for the Carot Notation codes. I.e.,:...
//
//	^@ ⇒ NUL (0x00)
//
//	^A ⇒ SOH (0x01)
//
//	^B ⇒ STX (0x02)
//
//	^C ⇒ ETX (0x03)
//
//	^D ⇒ EOT (0x04)
//
//	^E ⇒ ENQ (0x05)
//
//	^F ⇒ ACK (0x06)
//
//	^G ⇒ BEL (0x07)
//
//	^H ⇒ BS  (0x08)
//
//	^I ⇒ HT  (0x09)
//
//	^J ⇒ LF  (0x0a)
//
//	^K ⇒ VT  (0x0b)
//
//	^L ⇒ FF  (0x0c)
//
//	^M ⇒ CR  (0x0d)
//
//	^N ⇒ SO  (0x0e)
//
//	^O ⇒ SI  (0x0f)
//
//	^P ⇒ DLE (0x10)
//
//	^Q ⇒ DC1 (0x11)
//
//	^R ⇒ DC2 (0x12)
//
//	^S ⇒ DC3 (0x13)
//
//	^T ⇒ DC4 (0x14)
//
//	^U ⇒ NAK (0x15)
//
//	^V ⇒ SYN (0x16)
//
//	^W ⇒ ETB (0x17)
//
//	^X ⇒ CAN (0x18)
//
//	^Y ⇒ EM  (0x19)
//
//	^Z ⇒ SUB (0x1a)
//
//	^[ ⇒ ESC (0x1b)
//
//	^\ ⇒ FS  (0x1c)
//
//	^] ⇒ GS  (0x1d)
//
//	^^ ⇒ RS	 (0x1e)
//
//	^_ ⇒ US	 (0x1f)
//
//	^` ⇒ SP  (0x20)
//
//	^? ⇒ DEL (0x7f)
//
// Or, for some example Go code:
//
//	var code rune = 'G'
//	
//	value := decode(code)
//	
//	// value == '\a' == rune(0x07)
//
// Or, for anoterh example Go code:
//
//	var code rune = '['
//	
//	value := decode(code)
//	
//	// value == rune(0x1b)
func decode(r rune) rune {
	return r ^ 0x40
}
