# Copyright 2019 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

from __future__ import print_function

import sys
from google.auth.credentials import AnonymousCredentials
from google.cloud import storage


if len(sys.argv)-1 != 3:
    print('args: PROJECT BUCKET record|replay')
    sys.exit(1)
project = sys.argv[1]
bucket_name = sys.argv[2]
mode = sys.argv[3]

if mode == 'record':
    creds = None  # use default creds for demo purposes; not recommended
    client = storage.Client(project=project)
elif mode == 'replay':
    creds = AnonymousCredentials()
else:
    print('want record or replay')
    sys.exit(1)

client = storage.Client(project=project, credentials=creds)
bucket = client.get_bucket(bucket_name)
print('bucket %s created %s' %(bucket.id, bucket.time_created))
