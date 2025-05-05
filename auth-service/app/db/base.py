# This file ensures all models are imported before Base is used by Alembic or other tools.
# It imports the Base class and all model classes.

from app.db.base_class import Base
from app.db.models.user import User # I need to import my User model here

# Add imports for other models here as they are created, e.g.:
# from app.db.models.item import Item 