func lineEnd(b *buffer, p pos) pos { return pos{p.line, b.lineLen(p.line)} }
