import { describe, expect, it } from 'vitest';

import { decodeTextFileBuffer, formatEncodingName } from './textFileDecoder';

describe('text file decoder', () => {
  it('decodes utf-8 text', () => {
    const decoded = decodeTextFileBuffer(new TextEncoder().encode('第一章\n正文。').buffer);

    expect(decoded.encoding).toBe('utf-8');
    expect(decoded.text).toContain('第一章');
  });

  it('decodes utf-16le text with bom', () => {
    const body = encodeUTF16LE('第一章\n正文。');
    const bytes = new Uint8Array(body.length + 2);
    bytes.set([0xff, 0xfe], 0);
    bytes.set(body, 2);

    const decoded = decodeTextFileBuffer(bytes.buffer);

    expect(decoded.encoding).toBe('utf-16le');
    expect(decoded.text).toContain('第一章');
  });

  it('falls back to gb18030 for common gbk novel text', () => {
    const bytes = new Uint8Array([0xb5, 0xda, 0xd2, 0xbb, 0xd5, 0xc2, 0x0a, 0xd5, 0xfd, 0xce, 0xc4, 0xa1, 0xa3]);

    const decoded = decodeTextFileBuffer(bytes.buffer);

    expect(['gb18030', 'gbk']).toContain(decoded.encoding);
    expect(decoded.text).toBe('第一章\n正文。');
  });

  it('formats encoding names for user messages', () => {
    expect(formatEncodingName('gb18030')).toBe('GB18030');
    expect(formatEncodingName('utf-16le')).toBe('UTF-16 LE');
  });
});

function encodeUTF16LE(text: string) {
  const bytes = new Uint8Array(text.length * 2);
  for (let index = 0; index < text.length; index += 1) {
    const code = text.charCodeAt(index);
    bytes[index * 2] = code & 0xff;
    bytes[index * 2 + 1] = code >> 8;
  }
  return bytes;
}
