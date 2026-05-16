# -*- coding: utf-8 -*-
"""
Parse 系统架构师真题 .docx (with headers like 【2021下架构真题第NN题】) into exam-client JSON.

Requires: pip install python-docx
"""
from __future__ import annotations

import json
import re
import sys
from pathlib import Path

try:
    from docx import Document
except ImportError:
    print("请先安装: pip install python-docx", file=sys.stderr)
    sys.exit(1)

HEADER_RE = re.compile(r"【2021下架构真题第(\d+)题[^】]*】")
ANSWER_RE = re.compile(r"解答[：:]\s*答案选择\s*(.+?)(?:。|．|$)")
OPT_LEADER = re.compile(r"^([ABCD])([\.．]?)\s*(.*)$")


def paragraph_texts(doc_path: Path) -> list[str]:
    d = Document(str(doc_path))
    return [p.text.replace("\r", "").strip() for p in d.paragraphs]


def find_blocks(lines: list[str]) -> list[tuple[int, list[str]]]:
    starts: list[tuple[int, int]] = []
    for i, t in enumerate(lines):
        m = HEADER_RE.match(t.strip())
        if m:
            starts.append((i, int(m.group(1))))
    blocks: list[tuple[int, list[str]]] = []
    for idx, (pos, qn) in enumerate(starts):
        end = starts[idx + 1][0] if idx + 1 < len(starts) else len(lines)
        chunk = [lines[j] for j in range(pos + 1, end) if lines[j].strip()]
        blocks.append((qn, chunk))
    return blocks


def split_answer_letters(ans_blob: str) -> list[str]:
    ans_blob = ans_blob.strip()
    parts = re.split(r"[｜|]", ans_blob)
    letters: list[str] = []
    for p in parts:
        for ch in p:
            if ch in "ABCD":
                letters.append(ch)
                break
    return letters


def expand_jing_lines(lines: list[str]) -> list[str]:
    """Split inline fullwidth 》 so delimiter is its own line for option groups."""
    out: list[str] = []
    for ln in lines:
        if "》" not in ln:
            out.append(ln)
            continue
        parts = ln.split("》")
        for i, p in enumerate(parts):
            p = p.strip()
            if p:
                out.append(p)
            if i < len(parts) - 1:
                out.append("》")
    return out


def merge_option_lines(raw: list[str]) -> list[str]:
    """Normalize A. x / lone A + next line into 'A.body'."""
    lines = [x.strip() for x in raw if x.strip() and x.strip() != "》"]
    out: list[str] = []
    j = 0
    while j < len(lines):
        ln = lines[j]
        m = OPT_LEADER.match(ln)
        if m:
            letter, dot, rest = m.groups()
            if dot or rest:
                out.append(f"{letter}.{rest}".strip())
                j += 1
                continue
            if j + 1 < len(lines):
                nxt = lines[j + 1]
                if not OPT_LEADER.match(nxt) and not (len(nxt) == 1 and nxt in "ABCD"):
                    out.append(f"{letter}.{nxt}".strip())
                    j += 2
                    continue
            out.append(f"{letter}.")
            j += 1
            continue
        if out:
            out[-1] = out[-1] + " " + ln
        j += 1
    return out


def merged_to_option_texts(merged: list[str]) -> list[str]:
    texts: list[str] = []
    for s in merged:
        m = OPT_LEADER.match(s.strip())
        if not m:
            continue
        _, _, body = m.groups()
        texts.append(body.strip())
    return texts


def split_option_groups(opt_raw: list[str]) -> list[list[str]]:
    groups_lines: list[list[str]] = []
    cur: list[str] = []
    for ln in opt_raw:
        if ln.strip() == "》":
            if cur:
                groups_lines.append(cur)
                cur = []
            continue
        cur.append(ln)
    if cur:
        groups_lines.append(cur)
    return [merged_to_option_texts(merge_option_lines(g)) for g in groups_lines]


def parse_block(qn: int, chunk: list[str]) -> list[dict]:
    """Returns one or more question dicts (multi-blank => multiple rows)."""
    split_at = None
    for i, ln in enumerate(chunk):
        if ln.strip().startswith("解答"):
            split_at = i
            break
    if split_at is None:
        return []

    body = chunk[:split_at]
    tail = chunk[split_at]
    am = ANSWER_RE.search(tail.replace("\n", ""))
    if not am:
        return []
    letters = split_answer_letters(am.group(1))
    if not letters:
        return []

    head_lines = expand_jing_lines(body[:])
    opt_raw: list[str] = []
    stem_parts: list[str] = []
    seen_option = False
    for ln in head_lines:
        s = ln.strip()
        if not s:
            continue
        if not seen_option:
            if re.match(r"^\d{1,3}\.", s):
                stem_parts.append(re.sub(r"^\d{1,3}\.\s*", "", s))
                continue
            if OPT_LEADER.match(s) or (len(s) == 1 and s in "ABCD"):
                seen_option = True
                opt_raw.append(ln)
                continue
            stem_parts.append(s)
        else:
            opt_raw.append(ln)

    stem = " ".join(stem_parts).strip()
    if not stem:
        stem = f"（真题第 {qn} 题）"

    groups = split_option_groups(opt_raw)
    groups = [g for g in groups if len(g) >= 2]
    if not groups:
        return []

    if len(groups) == 1:
        opts = groups[0][:4]
        if len(opts) < 2:
            return []
        letter = letters[0]
        idx = ord(letter) - ord("A")
        ans = opts[idx] if 0 <= idx < len(opts) else letter
        if isinstance(ans, str) and not ans.strip():
            ans = letter
        return [
            _item(stem, opts, ans, (qn, 0)),
        ]

    out: list[dict] = []
    for gi, opts in enumerate(groups):
        opts = opts[:4]
        if len(opts) < 2:
            continue
        letter = letters[gi] if gi < len(letters) else ""
        idx = ord(letter) - ord("A") if letter else -1
        ans = opts[idx] if 0 <= idx < len(opts) else letter
        if isinstance(ans, str) and not ans.strip():
            ans = letter
        suffix = f"（第 {gi + 1} 空）" if len(groups) > 1 else ""
        out.append(_item(stem + suffix, opts, ans, (qn, gi)))
    return out


def _item(content: str, opts: list[str], answer: str, sort_key: tuple[int, int]) -> dict:
    return {
        "_sort": sort_key,
        "type": 1,
        "content": content.replace("\n", " ").strip(),
        "options": "\n".join(opts[:4]),
        "answer": answer,
        "score": 1,
        "difficulty": 2,
        "tags": "软考-系统架构师-2021下",
    }


def main() -> None:
    root = Path(__file__).resolve().parent
    doc = root / "2021年11月系统架构师真题（综合知识+答案解析）.docx"
    if not doc.exists():
        print("找不到文件:", doc, file=sys.stderr)
        sys.exit(1)

    lines = paragraph_texts(doc)
    blocks = find_blocks(lines)
    all_items: list[dict] = []
    for qn, chunk in blocks:
        for item in parse_block(qn, chunk):
            all_items.append(item)

    all_items.sort(key=lambda x: x["_sort"])
    for it in all_items:
        it.pop("_sort", None)

    out_path = root / "2021年11月系统架构师真题_综合知识_questions.json"
    out_path.write_text(json.dumps(all_items, ensure_ascii=False, indent=2), encoding="utf-8")

    sys.stdout.buffer.write(
        f"written: {out_path}\nquestions: {len(all_items)}\n".encode("utf-8")
    )


if __name__ == "__main__":
    main()
