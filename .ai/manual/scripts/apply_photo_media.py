"""Replace ONLY generated manual media with licensed photography-based examples."""
from pathlib import Path
import json,shutil,subprocess
B=Path(__file__).resolve().parents[1];RUN=B/'.local/runtime'
manifest=json.loads((B/'photo-sources.json').read_text(encoding='utf-8'))
scenes=sorted([x for x in manifest['sources']if x['kind']=='scene'],key=lambda x:x['index'])
portraits=sorted([x for x in manifest['sources']if x['kind']=='portrait'],key=lambda x:x['index'])
tool=RUN/'ffmpeg/ffmpeg.exe';archive=B/'.local/illustrated-media-before-photography';archive.mkdir(exist_ok=True)
for entry in scenes:
 i=entry['index'];scene=B/entry['local_file'];poster=B/entry['poster_file']
 for target,source in [(RUN/f'db/resCoverPoster/manual-{i:02d}.jpg',poster),(RUN/f'media/atlas/{i:02d}.jpg',scene),(RUN/f'media/comic/{i:02d}.jpg',poster)]:
  target.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(source,target)
 output=RUN/f'media/videos/CM-{i:03d}.mp4';stamp=RUN/f'media/videos/CM-{i:03d}.photo-source'
 if stamp.exists()and stamp.read_text()==entry['sha256']:continue
 if output.exists()and not(archive/output.name).exists():shutil.copyfile(output,archive/output.name)
 subprocess.run([str(tool),'-hide_banner','-loglevel','error','-y','-loop','1','-framerate','10','-i',str(scene),'-f','lavfi','-i',f'sine=frequency={220+(i-1)*12}:sample_rate=22050','-t','32','-vf','scale=640:426,format=yuv420p','-c:v','libx264','-preset','ultrafast','-crf','28','-c:a','aac','-b:a','32k','-movflags','+faststart',str(output)],check=True)
 stamp.write_text(entry['sha256']);print(f'photo video {i:02d} ready',flush=True)
for entry in portraits:shutil.copyfile(B/entry['local_file'],RUN/f"db/performerFace/manual-{entry['index']:02d}.jpg")
# Exact file copies provide honest, reproducible duplicate-detection examples.
# No fingerprint or success record is fabricated; the application computes it.
for i in range(8):
 source=RUN/f'media/videos/CM-{i+1:03d}.mp4';target=RUN/f'media/videos/CM-{i+17:03d}.mp4'
 import hashlib
 if hashlib.sha256(source.read_bytes()).digest()!=hashlib.sha256(target.read_bytes()).digest():shutil.copyfile(source,target)
(RUN/'photography-duplicate-examples.json').write_text(json.dumps({'pairs':[[i+1,i+17]for i in range(8)],'purpose':'8 real file-copy pairs for fingerprint demonstration'},indent=2),encoding='utf-8')
(RUN/'media/README.txt').write_text('使用手册演示资源：摄影素材来自 Unsplash / Pexels，来源和授权见 photo-sources.json。作品与人物名称为虚构；视频为照片与合成音制作的演示片段。\n',encoding='utf-8')
old=json.loads((RUN/'asset-manifest.json').read_text(encoding='utf-8'));old.update(source='Licensed photography; see .ai/manual/photo-sources.json',photography=True,media_generator='apply_photo_media.py')
(RUN/'asset-manifest.json').write_text(json.dumps(old,ensure_ascii=False,indent=2),encoding='utf-8')
# Saved bookmark covers deliberately survive source replacement in the app.
# This documentation-only migration archives the old generated-art JPEGs and
# removes just those generated cover files; the real cover endpoint rebuilds
# them from the new photo media without deleting bookmark records or IDs.
cover_root=(RUN/'db/videoMomentCovers').resolve();archive_covers=B/'.local/illustrated-moment-covers'
if not cover_root.is_relative_to(RUN.resolve()):raise RuntimeError('manual cover root escaped runtime')
for cover in cover_root.glob('*/*.jpg'):
 if not cover.resolve().is_relative_to(cover_root):raise RuntimeError('cover escaped isolated cache')
 saved=archive_covers/cover.parent.name/cover.name;saved.parent.mkdir(parents=True,exist_ok=True)
 if not saved.exists():shutil.copyfile(cover,saved)
 cover.unlink()
print('Photography media installed inside the isolated manual runtime.')
