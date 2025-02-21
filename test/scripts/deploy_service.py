import os
import zipfile
import subprocess
import requests
from pathlib import Path

def get_project_root():
    """Get the project root directory."""
    return Path(__file__).parent.parent.parent

def build_service():
    """Build the greet service."""
    root = get_project_root()
    service_dir = root / "test/services/greet-service"
    exe_path = service_dir / "greet-service.exe"
    
    print("Building service...")
    # Simpler build command - no need for GOOS/GOARCH since we're in Linux
    env = os.environ.copy()
    env["CGO_ENABLED"] = "0"  # Still good to disable CGO for container deployment
    
    result = subprocess.run(
        ["go", "build", "-o", "greet-service.exe", "./server"],
        cwd=service_dir,
        capture_output=True,
        text=True,
        env=env
    )
    
    if result.returncode != 0:
        print("Build failed:", result.stderr)
        return False
    
    print("Build successful")
    return True

def create_zip():
    """Create a zip file containing the executable and config."""
    root = get_project_root()
    service_dir = root / "test/services/greet-service"
    zip_path = service_dir / "service.zip"
    exe_path = service_dir / "greet-service.exe"
    config_path = service_dir / "server" / "config.toml"
    
    with zipfile.ZipFile(zip_path, "w") as zf:
        zf.write(exe_path, arcname="greet-service.exe")
        zf.write(config_path, arcname="config.toml")
    
    print(f"Created zip file at {zip_path}")
    return zip_path

def install_service(zip_path):
    """Install the service to Gonolith."""
    print("Installing service to Gonolith...")
    with open(zip_path, "rb") as f:
        response = requests.post("http://localhost:8080/install-service", data=f)
    
    if response.status_code == 200:
        print(f"Service installed successfully. Service ID: {response.text}")
        return True
    else:
        print(f"Installation failed with status {response.status_code}: {response.text}")
        return False

def main():
    if not build_service():
        return
    
    try:
        zip_path = create_zip()
        install_service(zip_path)
        # Clean up zip file
        os.remove(zip_path)
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    main()