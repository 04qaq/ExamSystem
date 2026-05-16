# -*- coding: utf-8 -*-
"""Extract single-choice questions from the RTF .doc into exam-client import JSON."""
from __future__ import annotations

import json
import re
import sys
from pathlib import Path

try:
    from striprtf.striprtf import rtf_to_text
except ImportError:
    print("请先安装: pip install striprtf", file=sys.stderr)
    sys.exit(1)


def strip_rtf_binary_chunks(rtf: str) -> str:
    """Remove \\binN<binary> segments so striprtf can read past embedded images."""
    out: list[str] = []
    i = 0
    while i < len(rtf):
        if rtf.startswith("\\bin", i) and i + 4 < len(rtf) and rtf[i + 4].isdigit():
            j = i + 4
            while j < len(rtf) and rtf[j].isdigit():
                j += 1
            n = int(rtf[i + 4 : j])
            i = j + n
            continue
        out.append(rtf[i])
        i += 1
    return "".join(out)


def rtf_to_plain(rtf_bytes: bytes) -> str:
    raw = rtf_bytes.decode("latin-1", errors="replace")
    raw = strip_rtf_binary_chunks(raw)
    return rtf_to_text(raw)


def normalize_spaces(s: str) -> str:
    s = s.replace("\xa0", " ").replace("\u3000", " ")
    s = re.sub(r"[ \t]+", " ", s)
    s = re.sub(r"\n{3,}", "\n\n", s)
    return s.strip()


def split_stem_opt_and_tail(lines: list[str]) -> tuple[str, str, list[str]] | None:
    """
    Stem lines until first line containing A．/B．/…;
    option lines until 4 opts parsed (with continuation lines that contain markers);
    remaining lines returned as tail for the next question's bridge.
    """
    lines = [ln.strip() for ln in lines if ln.strip()]
    stem_lines: list[str] = []
    opt_i: int | None = None
    for idx, ln in enumerate(lines):
        if re.search(r"(?<![A-Za-z0-9_])([ABCD])[．.]", ln):
            opt_i = idx
            break
        stem_lines.append(ln)
    if opt_i is None:
        return None

    acc: list[str] = []
    k = opt_i
    while k < len(lines):
        ln = lines[k]
        if re.match(r"^\d{1,3}[、，]", ln):
            break
        cand = acc + [ln]
        o = opts_from_blob(" ".join(cand))
        acc = cand
        k += 1
        if o and len(o) >= 4:
            if k < len(lines) and not re.search(r"[ABCD][．.]", lines[k]):
                break
            if k >= len(lines):
                break
    stem = " ".join(stem_lines).strip()
    opt_blob = " ".join(acc).strip()
    tail = lines[k:]
    return stem, opt_blob, tail


def opts_from_blob(opt_blob: str) -> list[str] | None:
    flat = re.sub(r"\s+", " ", opt_blob.strip())
    parts = re.split(r"(?=[ABCD][．.])", flat)
    opts: list[str] = []
    for p in parts:
        p = p.strip()
        if not p:
            continue
        om = re.match(r"^([ABCD])[．.](.+)$", p)
        if not om:
            continue
        opts.append(om.group(2).strip())
    if len(opts) < 2:
        return None
    return opts[:4]


def parse_answer_key(text: str) -> dict[int, str]:
    ans: dict[int, str] = {}
    for m in re.finditer(r"(?<!\d)(\d{1,3})\s*[、，]\s*([ABCD])(?!\w)", text):
        ans[int(m.group(1))] = m.group(2).upper()
    return ans


def extract_questions(text: str) -> list[dict]:
    text = normalize_spaces(text)
    answers_all = parse_answer_key(text)

    cut = -1
    for mk in ("参考答案", "答案解析", "试题解析", "[解析]"):
        idx = text.find(mk)
        if idx != -1:
            cut = idx if cut == -1 else min(cut, idx)

    body = text[:cut] if cut != -1 else text
    hdr = body.find("单项选择")
    if hdr != -1:
        nl = body.find("\n", hdr)
        body = body[nl + 1 :] if nl != -1 else body

    lines = body.splitlines()
    parsed: list[tuple[int, str, list[str]]] = []
    bridge = ""
    i = 0
    while i < len(lines):
        ln = lines[i].strip()
        m = re.match(r"^(\d{1,3})\s*[、，]\s*(.*)$", ln)
        if not m:
            if ln:
                bridge = (bridge + "\n" + ln).strip() if bridge else ln
            i += 1
            continue
        qno = int(m.group(1))
        rest_parts = [m.group(2)] if m.group(2).strip() else []
        i += 1
        while i < len(lines):
            ln2 = lines[i].strip()
            if re.match(r"^\d{1,3}\s*[、，]", ln2):
                break
            rest_parts.append(ln2)
            i += 1
        so = split_stem_opt_and_tail(rest_parts)
        if not so:
            bridge = ("\n".join(rest_parts) + ("\n" + bridge if bridge else "")).strip()
            continue
        stem_suffix, opt_blob, tail = so
        opts = opts_from_blob(opt_blob)
        if not opts:
            bridge = ("\n".join(rest_parts) + ("\n" + bridge if bridge else "")).strip()
            continue
        prev_bridge = bridge
        stem = " ".join(
            filter(None, [prev_bridge.replace("\n", " ").strip(), stem_suffix])
        ).strip()
        bridge = "\n".join(tail).strip() if tail else ""
        parsed.append((qno, stem, opts))

    parsed.sort(key=lambda x: x[0])
    seen: set[int] = set()
    uniq: list[tuple[int, str, list[str]]] = []
    for row in parsed:
        if row[0] in seen:
            continue
        seen.add(row[0])
        uniq.append(row)

    prev_stem = ""
    filled: list[tuple[int, str, list[str]]] = []
    for qno, stem, opts in uniq:
        if not stem.strip() or stem.startswith("（题号"):
            stem = prev_stem or stem
        else:
            prev_stem = stem
        filled.append((qno, stem, opts))

    out: list[dict] = []
    for qno, stem, opts in filled:
        letter = answers_all.get(qno)
        if letter and letter in "ABCD":
            idx = ord(letter) - ord("A")
            answer_text = opts[idx] if 0 <= idx < len(opts) else letter
        else:
            answer_text = ""

        stem_clean = stem.replace("\n", " ").strip()
        if not stem_clean:
            stem_clean = f"（题号 {qno}，请手工补题干）"

        out.append(
            {
                "type": 1,
                "content": stem_clean,
                "options": "\n".join(opts),
                "answer": answer_text,
                "score": 1,
                "difficulty": 2,
                "tags": "软考-软件设计师-模拟",
            }
        )
    return out


def main() -> None:
    root = Path(__file__).resolve().parent
    doc = root / "01中级软件设计师上午试题模拟+答案详解.doc"
    if not doc.exists():
        print("找不到文件:", doc, file=sys.stderr)
        sys.exit(1)

    plain = normalize_spaces(rtf_to_plain(doc.read_bytes()))
    (root / "_extracted_plain.txt").write_text(plain, encoding="utf-8")

    items = extract_questions(plain)
    out_path = root / "01中级软件设计师上午试题模拟_questions.json"
    out_path.write_text(json.dumps(items, ensure_ascii=False, indent=2), encoding="utf-8")

    missing = sum(1 for x in items if not x["answer"])
    sys.stdout.buffer.write(
        f"written: {out_path} ({len(items)} questions, {missing} without answer)\n".encode(
            "utf-8"
        )
    )


if __name__ == "__main__":
    main()
