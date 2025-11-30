"""Storage abstraction for document file storage.

Supports:
- LocalStorage: Filesystem storage for local development
- R2Storage: Cloudflare R2 storage for production (S3-compatible)
"""

import os
from abc import ABC, abstractmethod
from pathlib import Path

import aiofiles
import boto3


class StorageBackend(ABC):
    """Abstract base class for storage backends."""

    @abstractmethod
    async def upload(self, path: str, data: bytes, content_type: str) -> str:
        """Upload file, return storage path."""

    @abstractmethod
    async def download(self, path: str) -> bytes:
        """Download file content."""

    @abstractmethod
    async def delete(self, path: str) -> None:
        """Delete file."""

    @abstractmethod
    def get_url(self, path: str, expires_in: int = 3600) -> str:
        """Get download URL (signed for R2, file:// for local)."""

    @abstractmethod
    async def exists(self, path: str) -> bool:
        """Check if file exists."""


class LocalStorage(StorageBackend):
    """Filesystem storage for local development."""

    def __init__(self, base_path: str = "./storage/documents"):
        self.base_path = Path(base_path)
        self.base_path.mkdir(parents=True, exist_ok=True)

    async def upload(self, path: str, data: bytes, content_type: str) -> str:
        full_path = self.base_path / path
        full_path.parent.mkdir(parents=True, exist_ok=True)
        async with aiofiles.open(full_path, "wb") as f:
            await f.write(data)
        return path

    async def download(self, path: str) -> bytes:
        async with aiofiles.open(self.base_path / path, "rb") as f:
            return await f.read()

    async def delete(self, path: str) -> None:
        file_path = self.base_path / path
        if file_path.exists():
            file_path.unlink()

    def get_url(self, path: str, expires_in: int = 3600) -> str:
        return f"file://{(self.base_path / path).absolute()}"

    async def exists(self, path: str) -> bool:
        return (self.base_path / path).exists()


class R2Storage(StorageBackend):
    """Cloudflare R2 storage for production (S3-compatible)."""

    def __init__(
        self,
        account_id: str,
        access_key: str,
        secret_key: str,
        bucket: str,
    ):
        self.bucket = bucket
        self.client = boto3.client(
            "s3",
            endpoint_url=f"https://{account_id}.r2.cloudflarestorage.com",
            aws_access_key_id=access_key,
            aws_secret_access_key=secret_key,
        )

    async def upload(self, path: str, data: bytes, content_type: str) -> str:
        self.client.put_object(
            Bucket=self.bucket,
            Key=path,
            Body=data,
            ContentType=content_type,
        )
        return path

    async def download(self, path: str) -> bytes:
        response = self.client.get_object(Bucket=self.bucket, Key=path)
        return response["Body"].read()

    async def delete(self, path: str) -> None:
        self.client.delete_object(Bucket=self.bucket, Key=path)

    def get_url(self, path: str, expires_in: int = 3600) -> str:
        return self.client.generate_presigned_url(
            "get_object",
            Params={"Bucket": self.bucket, "Key": path},
            ExpiresIn=expires_in,
        )

    async def exists(self, path: str) -> bool:
        try:
            self.client.head_object(Bucket=self.bucket, Key=path)
            return True
        except self.client.exceptions.ClientError:
            return False


def get_storage() -> StorageBackend:
    """Factory function to get storage backend based on environment."""
    backend = os.getenv("STORAGE_BACKEND", "local")

    if backend == "r2":
        return R2Storage(
            account_id=os.environ["R2_ACCOUNT_ID"],
            access_key=os.environ["R2_ACCESS_KEY_ID"],
            secret_key=os.environ["R2_SECRET_ACCESS_KEY"],
            bucket=os.environ["R2_BUCKET_NAME"],
        )

    return LocalStorage()
