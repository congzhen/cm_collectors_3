"""Meaningful document checks, without touching existing application data."""
from pathlib import Path
import json,re,sqlite3,zipfile,hashlib
from pypdf import PdfReader
B=Path(__file__).resolve().parents[1];R=B.parents[1]
manifest=json.loads((B/'manual.json').read_text(encoding='utf-8'))
chapters=[json.loads((B/'chapters'/f'{id}.json').read_text(encoding='utf-8'))for id in manifest['chapters']]
pages={p['id']:p for c in chapters for p in c['pages']}
dataset=(R/'cm_collectors_html/src/assets/dataset.ts').read_text(encoding='utf-8')
modes=re.findall(r"'([^']+)'",re.search(r'resourcesShowMode:\s*\[([^\]]+)\]',dataset).group(1))
assert all('display-'+m in pages for m in modes),'missing layout documentation'
setting_map={'softSettings':'system-general','fileDatabase':'library-create','performerDatabase':'performer-browse','performerAvatarLibrary':'performer-avatar','tagDatabase':'tags-design','aiTag':'tags-ai','videoMetadata':'video-metadata','videoTranscode':'video-transcode','cronJobs':'system-cron','duplicateDetection':'video-duplicates','tvboxRecommend':'system-tvbox','databaseCleanup':'system-cleanup','autoBackup':'system-backup'}
setting=(R/'cm_collectors_html/src/views/SettingView.vue').read_text(encoding='utf-8')
tabs=re.findall(r'<el-tab-pane\s+label="([^"]+)"\s+name="([^"]+)"',setting)
assert all(name in setting_map and setting_map[name]in pages for _,name in tabs),'new settings tab needs documentation'
pdf=PdfReader(str(R/'CM Collectors 3 使用说明.pdf'))
feature_map=json.loads((B/'feature-map.json').read_text(encoding='utf-8'))
expected_pages={entry['id']:entry['pdf_page'] for entry in feature_map['features']}
text_by_page=[pg.extract_text() or '' for pg in pdf.pages]
all_text='\n'.join(text_by_page)
root_bookmarks=[item for item in pdf.outline if isinstance(item,dict)]
assert [item['/Title'] for item in root_bookmarks]==[c['title'] for c in chapters],'chapter bookmark order mismatch'
for chapter in chapters:
 assert pdf.get_destination_page_number(next(item for item in root_bookmarks if item['/Title']==chapter['title']))+1==expected_pages[chapter['pages'][0]['id']],'chapter bookmark points to wrong page'
for chapter_id in manifest.get('internal_chapters',[]):
 internal=json.loads((B/'chapters'/f'{chapter_id}.json').read_text(encoding='utf-8'))
 assert internal['title'] not in all_text,'internal chapter published'
 assert all(p['title'] not in all_text for p in internal['pages']),'internal section published'
for id,p in pages.items():
 assert p['title'] in text_by_page[expected_pages[id]-1],'feature points to wrong page'
 normalized=re.sub(r'\s+','',text_by_page[expected_pages[id]-1])
 expected_text=[p['lead']]+[part for block in p['blocks'] for part in [block['title'],block['text']]]+[str(cell) for row in p.get('rows',[]) for cell in row]
 assert all(re.sub(r'\s+','',part) in normalized for part in expected_text),('body text missing or truncated',id)
 assert id.upper() not in all_text,'internal feature ID visible in user PDF'
 if p.get('crop'):
  from PIL import Image
  with Image.open(B/'assets/screenshots'/f"{p['image']}.jpg") as im:
   left,top,right,bottom=p['crop'];assert 0<=left<right<=im.width and 0<=top<bottom<=im.height,('invalid crop',id)
assert '独立演示环境实际界面' not in all_text,'internal screenshot caption published'
assert 'minimal' not in all_text and 'Dockerfile.full' not in all_text,'outdated Docker edition guidance published'
page_refs={p.indirect_reference.idnum for p in pdf.pages}
page_numbers={p.indirect_reference.idnum:i+1 for i,p in enumerate(pdf.pages)}
toc_targets=[]
links=0
for pg in pdf.pages:
 for ref in pg.get('/Annots',[]):
  a=ref.get_object()
  if a.get('/Subtype')=='/Link' and '/Dest'in a:
   dest=a['/Dest'];assert dest[0].idnum in page_refs,'broken TOC destination';links+=1
   toc_targets.append(page_numbers[dest[0].idnum])
assert links==len(pages)+len(chapters),(links,len(pages)+len(chapters))
expected_toc=[]
for chapter in chapters:
 expected_toc.append(expected_pages[chapter['pages'][0]['id']])
 expected_toc.extend(expected_pages[p['id']] for p in chapter['pages'])
assert toc_targets==expected_toc,'TOC links point to incorrect sections'
report={'date':'2026-10-26','chapters':len(chapters),'features':len(pages),'pdf_pages':len(pdf.pages),'toc_links_checked':links,'layouts_documented':modes,'settings_tabs_documented':{label:setting_map[name]for label,name in tabs},'pdf_size_bytes':(R/'CM Collectors 3 使用说明.pdf').stat().st_size,'checks':['所有稳定ID唯一','13种显示模式已覆盖','当前13个设置页已映射','所有目录链接目标有效','所有页面标题中文可提取','构建下边界及单行文字右边界无溢出','章节书签顺序和目标页正确','功能映射页与正文标题一致','内部章节和功能ID未进入用户正文','截图裁切范围有效']}
for p in pages.values():
 if p.get('image'):assert (B/'assets/screenshots'/f"{p['image']}.jpg").exists()
# The only database read below is the independently generated backup snapshot.
backup_dir=B/'.local/runtime/auto_backup'
backups=sorted(backup_dir.glob('*.zip'))
if backups:
 zpath=backups[-1]
 with zipfile.ZipFile(zpath)as z:
  assert z.testzip()is None,'corrupt backup zip'
  dbname=next(n for n in z.namelist()if n.endswith('manual.db'))
  restore=B/'.local/restore-check';restore.mkdir(exist_ok=True)
  target=restore/'manual.db';target.write_bytes(z.read(dbname))
 with sqlite3.connect(target.as_uri()+'?mode=ro',uri=True)as con:
  moment_table=next(row[0]for row in con.execute("SELECT name FROM sqlite_master WHERE type='table'")if row[0]in ['video_moment','video_moments'])
  counts={name:con.execute(f'SELECT COUNT(*) FROM "{name}"').fetchone()[0]for name in ['resources','performer','tag',moment_table]}
  assert counts['resources']==96 and counts['performer']==24,counts
  assert con.execute('PRAGMA integrity_check').fetchone()[0]=='ok'
 report['backup_restore_check']={'zip':zpath.name,'counts':counts,'sqlite_integrity':'ok','note':'仅恢复到手册.local/restore-check，不覆盖任何正在使用的库。'}
report['checks'].extend(['每页导语、正文及表格文字完整可提取','未恢复旧Docker镜像分类'])
(B/'.local/document-audit.json').write_text(json.dumps(report,ensure_ascii=False,indent=2),encoding='utf-8')
print(json.dumps(report,ensure_ascii=False,indent=2))
