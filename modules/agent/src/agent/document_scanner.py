from pathlib import Path
import logging
import re
from pypdf import PdfReader

logger = logging.getLogger(__name__)


def _clean_text(text: str) -> str:
    if not text:
        return ""
    text = re.sub(r"\n\s*\n", "\n\n", text)
    text = re.sub(r"[ \t]+", " ", text)
    return text.strip()


def _read_pdf_text(pdf_path: Path) -> str:
    try:
        reader = PdfReader(str(pdf_path))
    except FileNotFoundError:
        logger.warning(f"PDF not found: {pdf_path}")
        return ""
    except Exception as e:
        logger.error(f"Could not open PDF {pdf_path}: {e}")
        return ""

    parts: list[str] = []
    for i, page in enumerate(getattr(reader, "pages", [])):
        try:
            t = page.extract_text() or ""
        except Exception as e:
            logger.warning(f"Failed to extract text from {pdf_path} page {i+1}: {e}")
            t = ""
        if t:
            parts.append(t)
    return _clean_text("\n".join(parts))


def load_documents(root: str) -> tuple:
    root = Path(root)

    resume_text = "[Resume not available]"
    rtxt = _read_pdf_text(root / "resume.pdf")
    if rtxt:
        resume_text = rtxt
        logger.info("Resume loaded")

    summary = "[Summary not available]"
    try:
        s = (root / "summary.txt").read_text(encoding="utf-8").strip()
    except FileNotFoundError:
        logger.warning(f"Summary not found: {root / 'summary.txt'}")
        s = ""
    except Exception as e:
        logger.error(f"Could not load summary: {e}")
        s = ""
    if s:
        summary = s
        logger.info("Summary loaded")

    cover_letters: list[str] = []
    for i in (1, 2):
        t = _read_pdf_text(root / f"coverletter{i}.pdf")
        if t:
            cover_letters.append(t)
            logger.info(f"Cover letter {i} loaded")
        if not t:
            logger.warning(f"Cover letter {i} missing or empty")

    return resume_text, summary, cover_letters
