"""Download fixed, attributed photography into the isolated manual asset cache.
Source JPEGs stay ignored; the tracked manifest makes every image reproducible.
"""
from pathlib import Path
from concurrent.futures import ThreadPoolExecutor,as_completed
from urllib.request import Request,urlopen
import json,hashlib,time,shutil
from PIL import Image,ImageDraw,ImageFont
B=Path(__file__).resolve().parents[1];CACHE=B/'.local/photography';CACHE.mkdir(parents=True,exist_ok=True)
LANDSCAPES=[10,11,12,13,14,15,16,17,18,19,28,29,37,76,82,49,53,54,57,62,83,84,70,85]
OLD=json.loads((B/'photo-sources.json').read_text(encoding='utf-8'))['sources']if(B/'photo-sources.json').exists()else[]
PORTRAITS=[
 (1165720,'Min An'),(20319703,'Yusuf Çelik'),(20369209,'Sabine Meier'),(8377134,'PNW Production'),
 (30004315,'Daniel & Hannah Snipes'),(33242318,'Achi Murusidze'),(20756240,'Emiliano Vittoriosi'),(20861525,'Manuel Gamboa'),
 (18108800,'Dynamic Wang'),(36701322,'Becca Correia'),(10919461,'MENG MEDIA'),(9775543,'SHVETS production'),
 (27165527,'Kenneth Surillo'),(40268639,'Thien Le Duy'),(9045566,'Fiqhy Firdaus'),(38677835,'Ernest Flowers'),
 (20781046,'YIMING TANG'),(7956491,'Artem Podrez'),(31268615,"Duy's House of Photo"),
 (35503892,'Dunmarx Photography'),(20603309,'BYB BYB'),(220453,'Pixabay'),(415829,'Pixabay'),(614810,'Simon Robben')]
def get(url):
 for attempt in range(3):
  try:
   with urlopen(Request(url,headers={'User-Agent':'CM-Collectors-Manual/1.0'}),timeout=25)as r:return r.read()
  except Exception:
   if attempt==2:raise
   time.sleep(1)
def job(kind,index,id,author=''):
 file=CACHE/f'{kind}-{index+1:02d}-{id}.jpg'
 old=next((x for x in OLD if x['kind']==kind and x['index']==index+1 and x['id']==str(id)),None)
 if not file.exists()and old and(B/old['local_file']).exists():shutil.copyfile(B/old['local_file'],file)
 if kind=='scene':
  meta_file=CACHE/f'picsum-{id}.json'
  if not meta_file.exists():meta_file.write_bytes(get(f'https://picsum.photos/id/{id}/info'))
  meta=json.loads(meta_file.read_text(encoding='utf-8'))
  url=f'https://picsum.photos/id/{id}/960/640.jpg'
  record={'provider':'Unsplash via Lorem Picsum','source_page':meta['url'],'author':meta['author'],'license':'Unsplash License','license_url':'https://unsplash.com/license','download_url':url,'id':str(id)}
  poster=CACHE/f'poster-{index+1:02d}-{id}.jpg';poster_url=f'https://picsum.photos/id/{id}/600/850.jpg'
  if not poster.exists():poster.write_bytes(get(poster_url))
  record.update(poster_file=poster.relative_to(B).as_posix(),poster_download_url=poster_url,poster_sha256=hashlib.sha256(poster.read_bytes()).hexdigest())
 else:
  url=f'https://images.pexels.com/photos/{id}/pexels-photo-{id}.jpeg?auto=compress&cs=tinysrgb&w=480&h=640&fit=crop'
  record={'provider':'Pexels','source_page':f'https://www.pexels.com/photo/{id}/','author':author,'license':'Pexels License','license_url':'https://www.pexels.com/license/','download_url':url,'id':str(id)}
 if not file.exists():file.write_bytes(get(url))
 with Image.open(file)as im:im.verify()
 record.update(kind=kind,index=index+1,local_file=file.relative_to(B).as_posix(),sha256=hashlib.sha256(file.read_bytes()).hexdigest(),bytes=file.stat().st_size)
 print(f'{kind}-{index+1:02d}: ready',flush=True);return record
def main():
 records=[]
 with ThreadPoolExecutor(max_workers=4)as pool:
  tasks=[pool.submit(job,'scene',i,id)for i,id in enumerate(LANDSCAPES)]+[pool.submit(job,'portrait',i,id,author)for i,(id,author)in enumerate(PORTRAITS)]
  for task in as_completed(tasks):records.append(task.result())
 records.sort(key=lambda x:(x['kind'],x['index']))
 manifest={'date':'2026-10-26','use':'仅用于软件使用手册演示；作品和人物名称为虚构，照片不表示真实身份、作品封皮或代言。','sources':records}
 (B/'photo-sources.json').write_text(json.dumps(manifest,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
 for kind in ['scene','portrait']:
  sheet=Image.new('RGB',(1200,720),'#e5ecec');draw=ImageDraw.Draw(sheet)
  for entry in [r for r in records if r['kind']==kind]:
   i=entry['index']-1;im=Image.open(B/entry['local_file']).convert('RGB');im.thumbnail((190,145));x=i%6*200+(200-im.width)//2;y=i//6*180
   sheet.paste(im,(x,y));draw.text((i%6*200+8,y+150),f"{entry['index']:02d} / {entry['id']}",fill='#183941')
  sheet.save(CACHE/f'contact-{kind}.jpg',quality=90)
 print(f'{len(records)} photographs downloaded; source manifest saved.')
if __name__=='__main__':main()
