import os
import zipfile
import subprocess
import requests
import hashlib
from pathlib import Path


def get_project_root():
    """Get the project root directory."""
    return Path(__file__).parent.parent.parent


def build_service():
    """Build the greet service for Linux while keeping .exe extension."""
    root = get_project_root()
    service_dir = root / "test/services/greet-service"
    exe_path = service_dir / "greet-service.exe"

    print("Building service for Linux...")
    env = os.environ.copy()
    env["GOOS"] = "linux"  # Set target OS to Linux
    env["GOARCH"] = "amd64"  # Set architecture
    env["CGO_ENABLED"] = "0"  # Disable cgo for static builds

    result = subprocess.run(
        [
            "go",
            "build",
            "-ldflags",
            "-extldflags '-static'",
            "-o",
            "greet-service.exe",
            "./server",
        ],
        cwd=service_dir,
        capture_output=True,
        text=True,
        env=env,
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

    print(f"Exe path: {exe_path} (exists: {exe_path.exists()})")
    print(f"Config path: {config_path} (exists: {config_path.exists()})")

    with zipfile.ZipFile(zip_path, "w") as zf:
        # Add executable with arcname to ensure correct name in zip
        if exe_path.exists():
            zf.write(exe_path, arcname="greet-service.exe")
        else:
            raise FileNotFoundError(f"Executable not found at {exe_path}")

        # Add config with arcname to ensure correct name in zip
        if config_path.exists():
            zf.write(config_path, arcname="config.toml")
        else:
            raise FileNotFoundError(f"Config not found at {config_path}")

    # Verify zip contents
    with zipfile.ZipFile(zip_path, "r") as zf:
        print("\nZip contents:")
        for info in zf.infolist():
            print(f"- {info.filename} (size: {info.file_size} bytes)")

    print(f"\nCreated zip file at {zip_path}")
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
        print(
            f"Installation failed with status {response.status_code}: {response.text}"
        )
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
