"""Read only source paths explicitly recorded in the manual feature map."""
from pathlib import Path
import hashlib,json
B=Path(__file__).resolve().parents[1];R=B.parents[1]
mapping=json.loads((B/'feature-map.json').read_text(encoding='utf-8'))
changes={}
for f in mapping['features']:
 for source,expected in f['source_hashes'].items():
  path=R/source
  actual=hashlib.sha256(path.read_bytes()).hexdigest()if path.is_file()else None
  if actual!=expected:changes.setdefault(source,[]).append(f['id'])
print(json.dumps({'changed_sources':changes,'note':'新功能还需检查新入口与新文件；仅未改变的已映射来源不代表全部功能覆盖。'},ensure_ascii=False,indent=2))
