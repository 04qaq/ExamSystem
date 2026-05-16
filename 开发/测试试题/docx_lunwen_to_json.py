# -*- coding: utf-8 -*-
"""
Parse 系统架构师「论文」真题 .docx into JSON (bulk-import compatible + structured topics).

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

TOPIC_START = re.compile(r"^试题(?P<num>[一二三四五六七八九十\d]+)[：:]\s*(?P<title>.+)$")
REQ_LINE = re.compile(r"^\s*\d+[、,．.\)]\s*")


def cn_or_digits_to_int(s: str) -> int:
    m = re.match(r"^(\d+)$", s.strip())
    if m:
        return int(m.group(1))
    cmap = {
        "零": 0,
        "一": 1,
        "二": 2,
        "三": 3,
        "四": 4,
        "五": 5,
        "六": 6,
        "七": 7,
        "八": 8,
        "九": 9,
        "十": 10,
    }
    s = s.strip()
    if s in cmap:
        return cmap[s]
    if s.startswith("十") and len(s) > 1:
        return 10 + cmap.get(s[1], 0)
    if len(s) == 2 and s[1] == "十":
        return cmap.get(s[0], 0) * 10
    return 0


def paragraph_texts(doc_path: Path) -> list[str]:
    d = Document(str(doc_path))
    return [p.text.replace("\r", "").strip() for p in d.paragraphs]


def split_topic_blocks(lines: list[str]) -> list[list[str]]:
    blocks: list[list[str]] = []
    cur: list[str] = []
    for ln in lines:
        if not ln:
            continue
        if TOPIC_START.match(ln):
            if cur:
                blocks.append(cur)
            cur = [ln]
        elif cur:
            cur.append(ln)
    if cur:
        blocks.append(cur)
    return blocks


def parse_topic(block: list[str]) -> dict:
    first = block[0]
    m = TOPIC_START.match(first)
    if not m:
        return {}
    num_zh = m.group("num")
    title = m.group("title").strip()
    topic_no = cn_or_digits_to_int(num_zh)
    body = block[1:]
    prompt_idx = None
    for i, ln in enumerate(body):
        if ln.strip().startswith("请围绕"):
            prompt_idx = i
            break
    if prompt_idx is None:
        background = "\n\n".join(body)
        instruction = ""
        requirements: list[str] = []
    else:
        background = "\n\n".join(body[:prompt_idx]).strip()
        instruction = body[prompt_idx].strip()
        requirements = []
        for ln in body[prompt_idx + 1 :]:
            s = ln.strip()
            if REQ_LINE.match(s):
                requirements.append(s)

    content_parts = [first]
    if background:
        content_parts.append(background)
    if instruction:
        content_parts.append(instruction)
    content_parts.extend(requirements)
    content = "\n\n".join(content_parts)

    return {
        "topic_no": topic_no,
        "topic_no_zh": num_zh,
        "heading": title,
        "full_title": first,
        "background": background,
        "instruction": instruction,
        "requirements": requirements,
        "content_import": content,
    }


def topics_to_import_items(topics: list[dict]) -> list[dict]:
    items = []
    for t in topics:
        items.append(
            {
                "type": 5,
                "content": t["content_import"],
                "options": "",
                "answer": "开放式·人工阅卷",
                "score": 75,
                "difficulty": 3,
                "tags": "软考-系统架构师-2021下-论文",
            }
        )
    return items


def main() -> None:
    root = Path(__file__).resolve().parent
    doc = root / "2021年11月系统架构师真题（论文）.docx"
    if not doc.exists():
        print("找不到文件:", doc, file=sys.stderr)
        sys.exit(1)

    lines = paragraph_texts(doc)
    non_empty = [ln for ln in lines if ln.strip()]
    doc_title = ""
    if non_empty and not TOPIC_START.match(non_empty[0]):
        doc_title = non_empty[0].strip()
        non_empty = non_empty[1:]

    blocks = split_topic_blocks(non_empty)
    topics = []
    for b in blocks:
        parsed = parse_topic(b)
        if parsed:
            topics.append(parsed)

    payload = {
        "meta": {
            "source_docx": doc.name,
            "document_title": doc_title,
            "topic_count": len(topics),
        },
        "topics": topics,
        "questions": topics_to_import_items(topics),
    }

    out_struct = root / "2021年11月系统架构师真题_论文.json"
    out_struct.write_text(json.dumps(payload, ensure_ascii=False, indent=2), encoding="utf-8")

    out_import = root / "2021年11月系统架构师真题_论文_questions.json"
    out_import.write_text(
        json.dumps(payload["questions"], ensure_ascii=False, indent=2),
        encoding="utf-8",
    )

    sys.stdout.buffer.write(
        f"written: {out_struct}\nwritten: {out_import}\ntopics: {len(topics)}\n".encode(
            "utf-8"
        )
    )


if __name__ == "__main__":
    main()
