import os
import sys
from logging.config import fileConfig

# I need to ensure the directory containing 'app' (the auth-service root)
# is on the Python path when Alembic runs.
# Explicitly adding the current working directory (which should be auth-service)
# is often the most reliable way.
CWD = os.getcwd()
if CWD not in sys.path:
    sys.path.insert(0, CWD)

# Now the imports should work
from sqlalchemy import engine_from_config
from sqlalchemy import pool
from alembic import context
from app.db.base import Base # This imports all models defined in app/db/models/
from app.core.config import settings


# this is the Alembic Config object, which provides
# access to the values within the .ini file in use.
config = context.config

# Interpret the config file for Python logging.
# This line sets up loggers basically.
if config.config_file_name is not None:
    fileConfig(config.config_file_name)

# add your model's MetaData object here
# for 'autogenerate' support
# from myapp import mymodel
# target_metadata = mymodel.Base.metadata
target_metadata = Base.metadata # Use the metadata from my Base class

# other values from the config, defined by the needs of env.py,
# can be acquired:-
# my_important_option = config.get_main_option("my_important_option")
# ... etc.

def get_url():
    """Helper function to return the database URL from settings."""
    return settings.DATABASE_URL

def run_migrations_offline() -> None:
    """Run migrations in 'offline' mode.

    This configures the context with just a URL
    and not an Engine, though an Engine is acceptable
    here as well.  By skipping the Engine creation
    we don't even need a DBAPI to be available.

    Calls to context.execute() here emit the given string to the
    script output.

    """
    # url = config.get_main_option("sqlalchemy.url") # Get URL from alembic.ini (original)
    url = get_url() # Get URL from app settings
    context.configure(
        url=url,
        target_metadata=target_metadata,
        literal_binds=True,
        dialect_opts={"paramstyle": "named"},
        # I should include the naming convention here too
        render_as_batch=True, # Recommended for SQLite, generally safe
        compare_type=True,    # Detect column type changes
        include_schemas=True, # If using multiple schemas
        # Use the same naming convention as in base_class.py
        # This might require defining the convention dict here or importing it
        # For simplicity, relying on metadata's convention for now, but might need explicit set.
        # naming_convention=Base.metadata.naming_convention
    )

    with context.begin_transaction():
        context.run_migrations()

def run_migrations_online() -> None:
    """Run migrations in 'online' mode.

    In this scenario we need to create an Engine
    and associate a connection with the context.

    """
    # This configuration section is needed for online mode.
    # It uses the 'sqlalchemy.url' from alembic.ini by default.
    # I will modify it to use the URL from my settings.
    configuration = config.get_section(config.config_main_section)
    configuration["sqlalchemy.url"] = get_url()

    connectable = engine_from_config(
        # config.get_section(config.config_main_section),
        configuration,
        prefix="sqlalchemy.",
        poolclass=pool.NullPool,
    )

    with connectable.connect() as connection:
        context.configure(
            connection=connection,
            target_metadata=target_metadata,
            # Match offline configuration options
            render_as_batch=True,
            compare_type=True,
            include_schemas=True,
            # naming_convention=Base.metadata.naming_convention
        )

        with context.begin_transaction():
            context.run_migrations()


if context.is_offline_mode():
    run_migrations_offline()
else:
    run_migrations_online() 