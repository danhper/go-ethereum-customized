import gzip


def open_file(filename: str, mode="rt", **kwargs):
    if filename.endswith(".gz"):
        return gzip.open(filename, mode, **kwargs)
    else:
        return open(filename, mode, **kwargs)
