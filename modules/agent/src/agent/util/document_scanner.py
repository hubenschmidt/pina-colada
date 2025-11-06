"""
Document scanner with aggressive text compression for better readability and token efficiency
"""

from pathlib import Path
import logging
import re
from pypdf import PdfReader

logger = logging.getLogger(__name__)


def _compress_text(text: str) -> str:
    """
    Aggressively compress text for better readability and token efficiency.

    Changes:
    - Multiple newlines -> single newline
    - Multiple spaces -> single space
    - Remove excessive whitespace
    - Normalize line breaks
    """
    if not text:
        return ""

    # Replace multiple newlines with single newline
    text = re.sub(r"\n\s*\n+", "\n", text)

    # Replace multiple spaces/tabs with single space
    text = re.sub(r"[ \t]+", " ", text)

    # Remove leading/trailing whitespace from each line
    lines = [line.strip() for line in text.split("\n")]

    # Filter out empty lines
    lines = [line for line in lines if line]

    # Join with single newline
    text = "\n".join(lines)

    return text.strip()


def _read_pdf_text(pdf_path: Path) -> str:
    """Read and compress PDF text"""
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

    # Join pages with space instead of newline for even more compression
    return _compress_text(" ".join(parts))


def _read_text_file(file_path: Path) -> str:
    """Read and compress text file"""
    try:
        text = file_path.read_text(encoding="utf-8").strip()
        return _compress_text(text)
    except FileNotFoundError:
        logger.warning(f"File not found: {file_path}")
        return ""
    except Exception as e:
        logger.error(f"Could not load {file_path}: {e}")
        return ""


def load_documents(root: str) -> tuple:
    """
    Load and compress all documents from the specified directory.

    Returns:
        tuple: (resume_text, summary, sample_answers, cover_letters)

    All text is aggressively compressed for:
    - Better readability in LangSmith
    - More efficient token usage
    - Cleaner context windows
    """
    root = Path(root)

    # Load resume
    resume_text = _read_pdf_text(root / "resume.pdf")
    if resume_text:
        logger.info(f"Resume loaded ({len(resume_text)} chars)")
    else:
        resume_text = "[Resume not available]"
        logger.warning("Resume not loaded")

    # Load summary
    summary = _read_text_file(root / "summary.txt")
    if summary:
        logger.info(f"Summary loaded ({len(summary)} chars)")
    else:
        summary = "[Summary not available]"
        logger.warning("Summary not loaded")

    # Load sample answers
    sample_answers = _read_text_file(root / "sample_answers.txt")
    if sample_answers:
        logger.info(f"Sample answers loaded ({len(sample_answers)} chars)")
    else:
        sample_answers = "[Sample answers not available]"
        logger.warning("Sample answers not loaded")

    # Load cover letters
    cover_letters: list[str] = []
    for i in (1, 2):
        cover_letter = _read_pdf_text(root / f"coverletter{i}.pdf")
        if cover_letter:
            cover_letters.append(cover_letter)
            logger.info(f"Cover letter {i} loaded ({len(cover_letter)} chars)")
        else:
            logger.warning(f"Cover letter {i} not loaded")

    return resume_text, summary, sample_answers, cover_letters
