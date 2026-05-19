#!/usr/bin/env python3
"""Generate a pbkdf2_sha256 hash for the Nomos users file.

Usage:
  python scripts/hash_password.py            # prompts for password
  python scripts/hash_password.py <password>
"""
import base64
import getpass
import hashlib
import secrets
import sys

ITERATIONS = 600_000


def hash_password(password: str) -> str:
    salt = secrets.token_bytes(16)
    derived = hashlib.pbkdf2_hmac("sha256", password.encode("utf-8"), salt, ITERATIONS)
    return (
        f"pbkdf2_sha256${ITERATIONS}$"
        f"{base64.b64encode(salt).decode('ascii')}$"
        f"{base64.b64encode(derived).decode('ascii')}"
    )


def main() -> None:
    if len(sys.argv) > 2:
        sys.exit("usage: hash_password.py [<password>]")
    if len(sys.argv) == 2:
        password = sys.argv[1]
    else:
        password = getpass.getpass("Password: ")
    if not password:
        sys.exit("password must not be empty")
    print(hash_password(password))


if __name__ == "__main__":
    main()
