#!/usr/bin/env python3
"""Call local Executor MCP with its private file-backed API key, without printing it."""
import argparse
import json
import os
from pathlib import Path
import urllib.request


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--tool', help='MCP tool name; omit to list tools')
    parser.add_argument('--input', help='JSON arguments file, required with --tool')
    args = parser.parse_args()
    directory = Path(os.environ.get('EXECUTOR_DATA_DIR', str(Path.home() / '.local/share/rdrop-executor')))
    key = os.environ.get('EXECUTOR_API_KEY') or json.loads((directory / 'keys.json').read_text())['apiKey']
    session = None

    def call(method, params, number):
        nonlocal session
        headers = {'Authorization': f'Bearer {key}', 'Accept': 'application/json, text/event-stream', 'Content-Type': 'application/json'}
        if session:
            headers['Mcp-Session-Id'] = session
        payload = {'jsonrpc': '2.0', 'id': number, 'method': method, 'params': params}
        request = urllib.request.Request('http://127.0.0.1:4312/mcp', data=json.dumps(payload).encode(), headers=headers)
        with urllib.request.urlopen(request, timeout=300) as response:
            session = response.headers.get('mcp-session-id') or session
            raw = response.read().decode()
        if raw.startswith('event:') or raw.startswith('data:'):
            messages = [json.loads(line[6:]) for line in raw.splitlines() if line.startswith('data: ')]
            message = next(value for value in messages if value.get('id') == number)
        else:
            message = json.loads(raw)
        if 'error' in message:
            raise RuntimeError(json.dumps(message['error']))
        return message['result']

    call('initialize', {'protocolVersion': '2025-03-26', 'capabilities': {}, 'clientInfo': {'name': 'raindrop-cli-setup', 'version': '0.2.0'}}, 1)
    if args.tool:
        if not args.input:
            parser.error('--input is required with --tool')
        result = call('tools/call', {'name': args.tool, 'arguments': json.loads(Path(args.input).read_text())}, 2)
    else:
        result = call('tools/list', {}, 2)
    print(json.dumps(result, indent=2))


if __name__ == '__main__':
    main()
