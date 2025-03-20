from pydantic import Field
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    # Service configuration
    SERVICE_NAME: str = Field(default="agent", description="Name of the service")
    SERVICE_PORT: int = Field(default=50051, description="Port to run the service on")
    SERVICE_HOST: str = Field(default="0.0.0.0", description="Host to bind the service to")
    
    # Environment configuration
    ENVIRONMENT: str = Field(default="development", description="Environment (development, staging, production)")
    
    # Consul configuration for service discovery
    CONSUL_HOST: str = Field(default="localhost", description="Consul host address")
    CONSUL_PORT: int = Field(default=8500, description="Consul port")
    
    # Encryption for sensitive data
    ENCRYPTION_KEY: str = Field(default="your-encryption-key", description="Key for encrypting sensitive data")
    
    # Configuration service connection
    CONFIG_SERVICE_NAME: str = Field(default="configuration", description="Configuration service name for discovery")
    
    class Config:
        env_file = ".env"
        env_file_encoding = "utf-8"
        case_sensitive = True


def get_settings() -> Settings:
    """Get application settings from environment variables."""
    return Settings()


settings = get_settings() 