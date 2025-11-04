import logging, sys
import os
import shutil
import textwrap


def configure_logging(level=logging.INFO) -> None:
    # Make this idempotent
    root = logging.getLogger()
    if getattr(root, "_configured_by_app", False):
        return

    for h in list(root.handlers):
        root.removeHandler(h)

    handler = logging.StreamHandler(stream=sys.stdout)
    formatter = logging.Formatter(
        "%(asctime)s %(levelname)s %(name)s:%(lineno)d - %(message)s"
    )
    handler.setFormatter(formatter)
    root.addHandler(handler)
    root.setLevel(level)
    root._configured_by_app = True  # sentinel

    # Quiet noisy libs if you want
    logging.getLogger("httpx").setLevel(logging.WARNING)
    logging.getLogger("uvicorn").setLevel(logging.INFO)
    logging.getLogger("uvicorn.access").setLevel(logging.WARNING)


def _console_width(default=100):
    # Works even when no TTY (common in Docker)
    try:
        return shutil.get_terminal_size().columns
    except OSError:
        return int(os.environ.get("COLUMNS", default))


def log_wrapped(logger, prefix, text, limit=None, suffix="..."):
    if limit is not None:
        text = (text[:limit] + suffix) if len(text) > limit else text
    width = max(_console_width() - len(prefix), 100)  # don't get too narrow
    wrapped = textwrap.fill(text, width=width, subsequent_indent=" " * len(prefix))
    logger.info("%s%s", prefix, wrapped)
