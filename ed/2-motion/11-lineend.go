func lineEnd(b *buffer, p pos) pos { return pos{p.line, len(b.lines[p.line])} }
