import os
import subprocess
import time
import signal
import sys
import requests
import zipfile
from pathlib import Path

def get_project_root():
    """Get the absolute path to the project root."""
    return Path(__file__).parent.parent.parent.absolute()

class GonolithCluster:
    def __init__(self, num_nodes=2):
        self.project_root = get_project_root()
        self.num_nodes = num_nodes
        self.nodes = []
        self.node_port_map = {}
        self.base_http_port = 8080
        self.base_grpc_port = 50051
        self.base_memberlist_port = 7946
    
    def initialize(self):
        """Initialize port assignments and create configuration"""
        for i in range(self.num_nodes):
            node_name = f"gonolith{i+1}"
            http_port = self.base_http_port + i
            grpc_port = self.base_grpc_port + i
            memberlist_port = self.base_memberlist_port + i
            
            self.node_port_map[node_name] = {
                "http_port": http_port,
                "grpc_port": grpc_port,
                "memberlist_port": memberlist_port
            }
    
    def start_cluster(self):
        # Generate cluster members string for each node
        for node_name, ports in self.node_port_map.items():
            other_nodes = []
            for other_name, other_ports in self.node_port_map.items():
                if other_name != node_name:
                    other_nodes.append(f"localhost:{other_ports['memberlist_port']}")
            
            cluster_members = ",".join(other_nodes)
            
            # Create command with the absolute path to main.go
            main_path = self.project_root / "cmd" / "main.go"
            
            if not main_path.exists():
                print(f"Error: {main_path} does not exist")
                continue
                
            cmd = [
                "go", "run", str(main_path),
            ]
            
            # Set environment variables
            env = os.environ.copy()
            env["CGO_ENABLED"] = "0"  # Disable CGO to avoid header issues
            env["HTTP_PORT"] = str(ports["http_port"])
            env["GRPC_PORT"] = str(ports["grpc_port"])
            env["NODE_NAME"] = node_name
            
            # Use the correct memberlist port for each node
            env["MEMBERLIST_PORT"] = str(ports["memberlist_port"])
            
            if cluster_members:
                env["CLUSTER_MEMBERS"] = cluster_members
            
            # Start process
            print(f"Starting {node_name}...")
            process = subprocess.Popen(cmd, env=env, cwd=str(self.project_root))
            
            self.nodes.append({
                "name": node_name, 
                "process": process, 
                "http_port": ports["http_port"]
            })
            print(f"Started {node_name} on HTTP:{ports['http_port']}, gRPC:{ports['grpc_port']}, Memberlist:{ports['memberlist_port']}")
            
            # Give it a moment to start up
            time.sleep(5)  # Increased wait time further
        
    def stop_cluster(self):
        """Stop all Gonolith nodes"""
        for node in self.nodes:
            print(f"Stopping {node['name']}...")
            node["process"].terminate()
        
        # Wait for processes to terminate
        for node in self.nodes:
            node["process"].wait()
    
    def deploy_service(self, service_path, node_name=None):
        """Deploy a service to a specific node or round-robin"""
        # If no node specified, choose the first one
        if not self.nodes:
            print("No nodes running to deploy service to")
            return False
            
        if node_name is None:
            node_name = self.nodes[0]["name"]
            
        # Find the node in our list
        target_node = None
        for node in self.nodes:
            if node["name"] == node_name:
                target_node = node
                break
                
        if not target_node:
            print(f"Node {node_name} not found")
            return False
        
        # Get the node's HTTP port
        http_port = target_node["http_port"]
        grpc_port = self.node_port_map[node_name]["grpc_port"]
        
        # Build service with the correct gRPC port
        full_service_path = self.project_root / service_path
        
        if not full_service_path.exists():
            print(f"Error: Service path {full_service_path} does not exist")
            return False
        
        # Update the service's config.toml with the correct port
        config_path = full_service_path / "server" / "config.toml"
        if config_path.exists():
            with open(config_path, "r") as f:
                config_content = f.read()
            
            # Update gRPC port in config
            updated_config = config_content.replace('port = ""', f'port = "{grpc_port}"')
            
            with open(config_path, "w") as f:
                f.write(updated_config)
        
        # Build the service
        print("Building service...")
        env = os.environ.copy()
        env["CGO_ENABLED"] = "0"  # Disable CGO to avoid header issues
        
        build_result = subprocess.run(
            ["go", "build", "-o", "greet-service.exe", "./server"],
            cwd=str(full_service_path),
            env=env,
            capture_output=True,
            text=True
        )
        
        if build_result.returncode != 0:
            print(f"Build failed: {build_result.stderr}")
            return False
            
        print("Build successful")
        
        # Create zip file
        print("Creating zip file...")
        zip_path = full_service_path / "service.zip"
        exe_path = full_service_path / "greet-service.exe"
        
        with zipfile.ZipFile(zip_path, "w") as zf:
            if exe_path.exists():
                zf.write(exe_path, arcname="greet-service.exe")
            else:
                print(f"Executable not found at {exe_path}")
                return False
                
            if config_path.exists():
                zf.write(config_path, arcname="config.toml")
            else:
                print(f"Config not found at {config_path}")
                return False
        
        # Verify node is still running
        try:
            check_response = requests.get(f"http://localhost:{http_port}/get-status", timeout=2)
            if check_response.status_code != 200:
                print(f"Node {node_name} is not responding correctly")
                return False
        except requests.RequestException as e:
            print(f"Could not connect to node {node_name}: {e}")
            return False
            
        # Install service
        print(f"Installing service to http://localhost:{http_port}/install-service...")
        try:
            with open(zip_path, "rb") as f:
                response = requests.post(
                    f"http://localhost:{http_port}/install-service", 
                    data=f,
                    timeout=10  # Longer timeout for installation
                )
            
            if response.status_code == 200:
                print(f"Service installed successfully. Service ID: {response.text}")
                # Clean up zip file
                os.remove(zip_path)
                return True
            else:
                print(f"Installation failed with status {response.status_code}: {response.text}")
                return False
        except requests.RequestException as e:
            print(f"Error connecting to node: {e}")
            return False

def main():
    cluster = GonolithCluster(num_nodes=2)
    cluster.initialize()
    
    try:
        cluster.start_cluster()
        
        # Deploy service
        cluster.deploy_service("test/services/greet-service", "gonolith1")
        
        print("\nCluster running. Press Ctrl+C to stop...")
        while True:
            time.sleep(1)
    except KeyboardInterrupt:
        print("\nShutting down cluster...")
    finally:
        cluster.stop_cluster()

if __name__ == "__main__":
    main()