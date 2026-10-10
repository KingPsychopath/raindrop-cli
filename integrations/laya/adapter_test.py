"""Verify the adapter protocol without installing Laya or downloading weights."""
import io
import json
import sys
import types
import unittest
from unittest.mock import patch
import classify

class Router:
    def __init__(self):
        print("runtime log")
    def predict(self, state, questions, **kwargs):
        assert state == {"title": "Docs"}
        assert kwargs == {"model": "english"}
        return {"routing": {"model": "english"}, "answers": {
            "42": {"choice": "docs", "confidence": 0.9,
                   "probabilities": {"docs": 0.9, "needs_review": 0.1}, "abstention": True}}}

class AdapterTest(unittest.TestCase):
    def test_protocol(self):
        source = io.StringIO(json.dumps({"state": {"title": "Docs"}, "questions": {"42": {}}, "model": "english"}))
        output, logs = io.StringIO(), io.StringIO()
        with patch.dict(sys.modules, {"laya": types.SimpleNamespace(Router=Router)}), patch.object(sys, "stdin", source), patch.object(sys, "stdout", output), patch.object(sys, "stderr", logs):
            classify.main()
        answer = json.loads(output.getvalue())["answers"]["42"]
        self.assertEqual(answer["type"], "choice")
        self.assertTrue(answer["abstention"])
        self.assertIn("runtime log", logs.getvalue())

if __name__ == "__main__":
    unittest.main()
