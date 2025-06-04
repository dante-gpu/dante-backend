# This file ensures all models are imported before Base is used by Alembic or other tools.
# It imports the Base class and all model classes.

from app.db.base_class import Base  # noqa

# Import models that User depends on first, or that are more standalone
from app.db.models.profile import UserProfile # noqa
from app.db.models.rbac import Role, Permission # noqa 
from app.db.models.security import UserApiKey, LoginHistory # noqa

# Import User model last as it refers to the above
from app.db.models.user import User  # noqa

# Add imports for other models here as they are created, e.g.:
# from app.db.models.item import Item 