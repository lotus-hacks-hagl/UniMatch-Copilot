import os
import subprocess
import time
from pathlib import Path

def test_cmd(cmd_list, env_dict):
    print(f"--- Testing {' '.join(cmd_list)} ---")
    try:
        p = subprocess.Popen(cmd_list, env=env_dict, stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    except Exception as e:
        print(f"Failed to start: {e}")
        return

    # Wait for a bit
    time.sleep(2)
    poll = p.poll()
    if poll is not None:
        print(f"CRASHED! Exit code: {poll}")
        out, err = p.communicate()
        print(f"Stdout:\n{out.decode()}")
        print(f"Stderr:\n{err.decode()}")
    else:
        print(f"RUNNING FINE. Terminating it...")
        # send EOF to stdin to tell it to gracefully exit
        p.kill()
        out, err = p.communicate(timeout=3)

if __name__ == "__main__":
    env_base = os.environ.copy()

    skill_script = Path(".claude/skills/tinyfish-web-agent/scripts/extract.sh")
    print(f"TinyFish skill script exists: {skill_script.exists()} ({skill_script})")
    print(f"TINYFISH_API_KEY configured: {bool(env_base.get('TINYFISH_API_KEY'))}")

    # Test mcp-neo4j-cypher
    env_neo4j = env_base.copy()
    test_cmd(["uvx", "mcp-neo4j-cypher", "--transport", "stdio"], env_neo4j)
