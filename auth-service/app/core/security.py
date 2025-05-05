from passlib.context import CryptContext

# I need to create a CryptContext instance specifying the hashing algorithm.
# bcrypt is a strong and recommended choice.
# "auto" will use the default scheme (bcrypt) for hashing and automatically
# identify hashes of all supported schemes for verification.
pwd_context = CryptContext(schemes=["bcrypt"], deprecated="auto")

def verify_password(plain_password: str, hashed_password: str) -> bool:
    """I need a function to verify a plain password against a stored hash."""
    return pwd_context.verify(plain_password, hashed_password)

def get_password_hash(password: str) -> str:
    """I need a function to hash a plain password."""
    return pwd_context.hash(password) 