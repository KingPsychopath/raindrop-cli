#!/usr/bin/env python3
"""Real local-model transport check against a disposable, loopback Raindrop fixture.

Requires running model servers. Never uses or changes a real Raindrop account.
"""
import argparse
import copy
import json
import os
from pathlib import Path
import subprocess
import threading
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--binary', default='./bin/rdrop')
    parser.add_argument('--laya-url', default='http://127.0.0.1:8000/v1')
    parser.add_argument('--model-url', default='http://127.0.0.1:11434/v1')
    parser.add_argument('--model', default='qwen3:4b')
    parser.add_argument('--output', default='.artifacts/e2e/models')
    parser.add_argument('--command-python', help='Also test the Laya command adapter with this Python')
    args = parser.parse_args()
    output = Path(args.output).resolve()
    output.mkdir(parents=True, exist_ok=True)
    items = {
        42: {'_id': 42, 'title': 'Go net/http API reference documentation', 'link': 'https://pkg.go.dev/net/http', 'excerpt': 'Official technical reference for HTTP clients and servers in Go.', 'note': 'Fixture private note', 'collection': {'$id': -1}, 'tags': ['personal']},
        43: {'_id': 43, 'title': 'Pasta recipe', 'link': 'https://fixture.example/pasta', 'excerpt': 'Ingredients, cooking times and instructions for homemade pasta.', 'collection': {'$id': -1}, 'tags': ['cooking']},
    }
    requests = []
    writes = []

    class Fixture(BaseHTTPRequestHandler):
        def log_message(self, *unused):
            pass

        def send(self, value, status=200):
            data = json.dumps(value).encode()
            self.send_response(status)
            self.send_header('Content-Type', 'application/json')
            self.send_header('Content-Length', str(len(data)))
            self.end_headers()
            self.wfile.write(data)

        def do_GET(self):
            route = self.path.split('?')[0]
            requests.append({'method': 'GET', 'path': self.path})
            if route == '/user':
                return self.send({'user': {'_id': 7, 'name': 'Disposable model fixture'}})
            if route == '/user/stats':
                return self.send({'items': [{'_id': -1, 'count': 2}]})
            if route == '/filters/0':
                return self.send({'result': True, 'notag': {'count': 0}, 'duplicates': {'count': 0}, 'broken': {'count': 0}, 'tags': [{'tag': 'reference', 'count': 1}]})
            if route == '/raindrops/0':
                return self.send({'result': True, 'items': list(items.values()), 'count': 2})
            if route in ['/collections', '/collections/childrens']:
                return self.send({'result': True, 'items': [{'_id': 10, 'title': 'Reference', 'count': 0}] if route == '/collections' else []})
            if route == '/tags/0':
                return self.send({'result': True, 'items': [{'_id': name, 'count': 1} for name in ['personal', 'reference', 'cooking']]})
            if route.startswith('/raindrop/') and int(route.rsplit('/', 1)[1]) in items:
                return self.send({'result': True, 'item': items[int(route.rsplit('/', 1)[1])]})
            return self.send({'error': 'Unexpected fixture route'}, 404)

        def do_PUT(self):
            route = self.path.split('?')[0]
            if not route.startswith('/raindrop/'):
                return self.send({'error': 'Unexpected fixture write'}, 404)
            bookmark_id = int(route.rsplit('/', 1)[1])
            body = json.loads(self.rfile.read(int(self.headers['Content-Length'])))
            writes.append({'id': bookmark_id, 'body': body})
            items[bookmark_id].update(body)
            return self.send({'result': True, 'item': items[bookmark_id]})

    server = ThreadingHTTPServer(('127.0.0.1', 0), Fixture)
    threading.Thread(target=server.serve_forever, daemon=True).start()
    environment = os.environ.copy()
    environment.update({
        'RAINDROP_API_BASE_URL': f'http://127.0.0.1:{server.server_port}',
        'RAINDROP_TOKEN': 'disposable-fixture-token',
        'RDROP_DECISION_PROVIDER': 'systemone',
        'RDROP_DECISION_BASE_URL': args.laya_url,
        'RDROP_DECISION_MODEL': 'english',
    })
    for key in ['RDROP_MODEL_API_KEY', 'OPENAI_API_KEY', 'RDROP_DECISION_API_KEY', 'TYPESAFE_API_KEY']:
        environment.pop(key, None)

    def run(name, arguments, extra=None):
        result = subprocess.run([str(Path(args.binary).resolve()), *arguments], env={**environment, **(extra or {})}, capture_output=True, text=True, timeout=900)
        (output / f'{name}.json').write_text(result.stdout)
        (output / f'{name}.stderr.txt').write_text(result.stderr)
        if result.returncode:
            raise RuntimeError(f'{name} failed: {result.stderr}; inspect {output}')
        return json.loads(result.stdout)

    try:
        taxonomy = output / 'taxonomy.json'
        taxonomy.write_text(json.dumps({'reference': 'Technical API reference documentation and engineering guides', 'cooking': 'Food recipes, ingredients and cooking instructions'}))
        run('report', ['cleanup', 'report', '--json'])
        classification = run('laya-http', ['classify', '--ids', '42,43', '--taxonomy', str(taxonomy), '--json'])
        assert len(classification['results']) == 2
        assert [(r['id'], r['category']) for r in classification['results']] == [(42, 'reference'), (43, 'cooking')]
        if args.command_python:
            command = run('laya-command', ['classify', '--ids', '42,43', '--taxonomy', str(taxonomy), '--json'], {
                'RDROP_DECISION_PROVIDER': 'command', 'RDROP_DECISION_COMMAND': str(Path(args.command_python).absolute()),
                'RDROP_DECISION_ARGS': json.dumps([str(Path('integrations/laya/classify.py').resolve())]),
            })
            assert [(r['id'], r['category']) for r in command['results']] == [(r['id'], r['category']) for r in classification['results']]
        task = ('/no_think This is a disposable test library. First call raindrop_classify for ids [42] '
                'with categories reference="Technical API reference documentation and engineering guides" and '
                'cooking="Food recipes and cooking instructions". Wait for its result before planning. '
                'Then call raindrop_organize_plan using the exact digest from the initial library context '
                'and a suggestion for id 42 with tags ["reference"], collectionId 10, reason "API reference", '
                'and needsReview copied from the classification. This is a read-only plan explicitly '
                'requested for the transport test. Do not apply anything. After both tools succeed, '
                'summarize their actual results and any classification uncertainty in one paragraph.')
        agent = run('agent', ['agent', '--base-url', args.model_url, '--api', 'chat', '--model', args.model,
                              '--max-steps', '8', '--max-tool-calls', '12', '--max-output-tokens', '4096', '--json', task])
        assert agent['status'] == 'complete' and agent['response']
        tool_names = {entry['tool'] for entry in agent['toolResults'] if not entry['isError']}
        assert {'raindrop_classify', 'raindrop_organize_plan'} <= tool_names, tool_names
        assert not writes, 'Read-only loop mutated the fixture'
        assert agent['plans'], 'No exact preview plan was captured'
        plan = output / 'reviewed-plan.json'
        plan.write_text(json.dumps(agent['plans'][0]))
        run('dry-run', ['cleanup', 'apply', str(plan), '--dry-run'])
        assert not writes, 'Dry-run mutated the fixture'
        before = copy.deepcopy(items[42])
        applied = run('apply-fixture', ['cleanup', 'apply', str(plan), '--yes'])
        assert len(writes) == 1 and items[42]['tags'] == ['personal', 'reference']
        assert items[42]['collection'] == {'$id': 10}
        assert items[42]['note'] == before['note']
        summary = {'status': 'passed', 'model': args.model, 'modelSteps': agent['steps'], 'modelToolCalls': agent['toolCalls'],
                   'tools': sorted(tool_names), 'classification': classification, 'fixtureWrites': len(writes),
                   'realRaindropWrites': 0, 'apply': applied}
        (output / 'summary.json').write_text(json.dumps(summary, indent=2) + '\n')
        print(json.dumps(summary, indent=2))
    finally:
        (output / 'fixture-requests.json').write_text(json.dumps(requests, indent=2) + '\n')
        server.shutdown()


if __name__ == '__main__':
    main()
