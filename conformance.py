#!/usr/bin/env python3

import pathlib
import os
import shlex
import subprocess

results = 'conformance_results'
root = 'examples/v1beta1'
crds = ['pipelineruns', 'taskruns']
dir_path = pathlib.Path(__file__).parent.absolute()

tests = []
for crd in crds:
    p = os.path.join(dir_path, root, crd)
    tests = tests + [(crd, os.path.splitext(f)[0]) for f in os.listdir(p) if os.path.isfile(os.path.join(p, f))]

f = open('conformance_results', 'w', buffering=1)
for t in tests:
    crd, test = t
    command = "go test -v -count=1 -tags=examples ./test/ -run \"TestExamples/v1beta1/{}/{}\"".format(crd, test)
    try:
        print("running: {}", command)
        subprocess.run(shlex.split(command), check=True)
    except subprocess.CalledProcessError as err:
        result = "{} {} failed ({})".format(crd, test, command)
    else:
        result = "{} {} passed".format(crd, test)

    f.write("{}\n".format(result))


