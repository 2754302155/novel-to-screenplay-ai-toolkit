export interface DecodedTextFile {
  text: string;
  encoding: string;
}

const fallbackEncodings = ['gb18030', 'gbk', 'big5'];

export function decodeTextFileBuffer(buffer: ArrayBuffer): DecodedTextFile {
  const bytes = new Uint8Array(buffer);
  const bomEncoding = detectBOM(bytes);
  if (bomEncoding) {
    return decodeWithEncoding(bytes, bomEncoding);
  }

  const utf16Encoding = detectUTF16ByNullBytes(bytes);
  if (utf16Encoding) {
    return decodeWithEncoding(bytes, utf16Encoding);
  }

  const utf8Text = tryDecode(bytes, 'utf-8', true);
  if (utf8Text !== null) {
    return { text: normalizeDecodedText(utf8Text), encoding: 'utf-8' };
  }

  return chooseReadableFallback(bytes);
}

export function formatEncodingName(encoding: string) {
  const normalized = encoding.toLowerCase();
  if (normalized === 'utf-8') {
    return 'UTF-8';
  }
  if (normalized === 'utf-16le') {
    return 'UTF-16 LE';
  }
  if (normalized === 'utf-16be') {
    return 'UTF-16 BE';
  }
  if (normalized === 'gb18030') {
    return 'GB18030';
  }
  if (normalized === 'gbk') {
    return 'GBK';
  }
  if (normalized === 'big5') {
    return 'Big5';
  }
  return encoding.toUpperCase();
}

function detectBOM(bytes: Uint8Array) {
  if (bytes.length >= 3 && bytes[0] === 0xef && bytes[1] === 0xbb && bytes[2] === 0xbf) {
    return 'utf-8';
  }
  if (bytes.length >= 2 && bytes[0] === 0xff && bytes[1] === 0xfe) {
    return 'utf-16le';
  }
  if (bytes.length >= 2 && bytes[0] === 0xfe && bytes[1] === 0xff) {
    return 'utf-16be';
  }
  return '';
}

function detectUTF16ByNullBytes(bytes: Uint8Array) {
  const sampleLength = Math.min(bytes.length, 200);
  if (sampleLength < 20) {
    return '';
  }

  let evenNulls = 0;
  let oddNulls = 0;
  for (let index = 0; index < sampleLength; index += 1) {
    if (bytes[index] !== 0) {
      continue;
    }
    if (index % 2 === 0) {
      evenNulls += 1;
    } else {
      oddNulls += 1;
    }
  }

  const pairs = Math.floor(sampleLength / 2);
  if (oddNulls / pairs > 0.35 && evenNulls / pairs < 0.05) {
    return 'utf-16le';
  }
  if (evenNulls / pairs > 0.35 && oddNulls / pairs < 0.05) {
    return 'utf-16be';
  }
  return '';
}

function decodeWithEncoding(bytes: Uint8Array, encoding: string): DecodedTextFile {
  return {
    text: normalizeDecodedText(new TextDecoder(encoding).decode(bytes)),
    encoding
  };
}

function chooseReadableFallback(bytes: Uint8Array): DecodedTextFile {
  const candidates = fallbackEncodings
    .map((encoding) => {
      const text = tryDecode(bytes, encoding, false);
      return text === null
        ? null
        : {
            text: normalizeDecodedText(text),
            encoding,
            score: scoreReadableChineseText(text)
          };
    })
    .filter((candidate): candidate is DecodedTextFile & { score: number } => candidate !== null)
    .sort((left, right) => right.score - left.score);

  if (candidates.length > 0) {
    const { text, encoding } = candidates[0];
    return { text, encoding };
  }

  return {
    text: normalizeDecodedText(new TextDecoder('utf-8').decode(bytes)),
    encoding: 'utf-8'
  };
}

function tryDecode(bytes: Uint8Array, encoding: string, fatal: boolean) {
  try {
    return new TextDecoder(encoding, { fatal }).decode(bytes);
  } catch {
    return null;
  }
}

function normalizeDecodedText(text: string) {
  return text.replace(/^\uFEFF/, '');
}

function scoreReadableChineseText(text: string) {
  let score = 0;
  for (const character of text) {
    const codePoint = character.codePointAt(0) ?? 0;
    if (isCJK(codePoint)) {
      score += 3;
    } else if (/[，。！？；：“”‘’、（）《》]/u.test(character)) {
      score += 2;
    } else if (character === '\uFFFD') {
      score -= 30;
    } else if (codePoint < 32 && !['\n', '\r', '\t'].includes(character)) {
      score -= 10;
    }
  }
  return score;
}

function isCJK(codePoint: number) {
  return (
    (codePoint >= 0x3400 && codePoint <= 0x4dbf) ||
    (codePoint >= 0x4e00 && codePoint <= 0x9fff) ||
    (codePoint >= 0xf900 && codePoint <= 0xfaff)
  );
}
