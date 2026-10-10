#!/usr/bin/env python3
"""Optional Laya command adapter: one System One JSON request in, answer out.
Install Laya separately. First use may download the selected checkpoint.
"""
import contextlib
import json
import sys


def main():
    request = json.load(sys.stdin)
    # Keep imports, downloads and runtime logs out of the JSON protocol stream.
    with contextlib.redirect_stdout(sys.stderr):
        from laya import Router
        router = Router()
        result = router.predict(request["state"], request["questions"], **(
            {"model": request["model"]} if request.get("model") else {}))
    answers = {}
    for key, value in result["answers"].items():
        answers[key] = {**value, "type": "choice"}
    json.dump({"model": result.get("routing", {}).get("model", request.get("model", "laya")),
               "answers": answers}, sys.stdout, allow_nan=False)
    sys.stdout.write("\n")


if __name__ == "__main__":
    main()
