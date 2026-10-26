"""Original, small, repeatable assets. No reads of existing app data/config/builds."""
from pathlib import Path
import json, math, random, wave, struct
from PIL import Image, ImageDraw, ImageFont

BASE = Path(__file__).resolve().parents[1]
RUN = BASE / '.local/runtime'
FONT = 'C:/Windows/Fonts/msyh.ttc'
TITLES = ['山海之间','晨光森林','云端漫步','城市回声','海岸来信','星河旅人','秋日山谷','雨后小镇','极地之光','湖畔日记','风的方向','远方灯塔','青山有约','落日航线','春日花园','沙丘足迹','雪岭清晨','河流记忆','夜色港湾','旷野之歌','岛屿时光','流云手记','竹林听雨','南方车站']

def scenic(file, index, title, size=(480,680)):
    rng = random.Random(3000+index)
    w,h=size; im=Image.new('RGB',size); d=ImageDraw.Draw(im)
    palettes=[((23,64,86),(96,171,176),(234,191,130)),((45,48,82),(126,138,176),(249,191,149)),((32,80,73),(129,180,151),(247,215,156)),((103,66,72),(208,141,122),(252,223,169))]
    sky,mid,sun=palettes[index%4]
    for y in range(h):
        t=y/h; color=tuple(int(sky[k]*(1-t)+mid[k]*t) for k in range(3)); d.line((0,y,w,y),fill=color)
    x=int(w*(.25+.5*rng.random())); y=int(h*.27); r=w//10
    d.ellipse((x-r,y-r,x+r,y+r),fill=sun)
    for layer in range(4):
        points=[(0,h)]+[(xx,int(h*(.46+layer*.12)+math.sin(xx/w*7+index+layer)*h*.09+rng.random()*h*.04)) for xx in range(0,w+20,20)]+[(w,h)]
        d.polygon(points, fill=tuple(max(0,c-int(layer*15)) for c in sky))
    d.text((w*.08,h*.70),'CM / ORIGINAL COLLECTION',font=ImageFont.truetype(FONT,max(10,w//30)),fill=(220,229,225))
    d.text((w*.08,h*.77),title,font=ImageFont.truetype(FONT,w//13),fill='white')
    d.text((w*.08,h*.89),f'VOL. {index+1:02d}    |    2026',font=ImageFont.truetype(FONT,w//28),fill=(220,229,225))
    file.parent.mkdir(parents=True,exist_ok=True); im.save(file,quality=83,optimize=True)

def main():
    if (RUN/'photography-applied.json').exists():
        print('Existing photography demo preserved; use apply_photo_media.py to refresh licensed media.')
        return
    RUN.mkdir(parents=True,exist_ok=True)
    for directory in ['db/resCoverPoster','db/performerFace','media/videos','media/atlas','media/comic','cert']:
        (RUN/directory).mkdir(parents=True,exist_ok=True)
    config='''general:
  logoName: CM 演示资料馆
  theme: bright
  homeMode: classic
  isAdminLogin: false
system:
  serverHost: 127.0.0.1
  port: 18345
  database: sqlite3
  filePath: ./db/
  Env: release
  logFilePath: ./manual-error.log
sqlite3:
  path: ./db/manual.db
  loggerLevel: silent
serverFileManagement:
  rootPath:
    - "MEDIA_ROOT"
performerAvatarLibrary:
  cachePath: ./runtime/cache/gfriends/
autoBackup:
  enabled: false
'''.replace('MEDIA_ROOT',(RUN/'media').as_posix())
    if not (RUN/'config.yaml').exists(): (RUN/'config.yaml').write_text(config,encoding='utf-8')
    for i,title in enumerate(TITLES):
        scenic(RUN/f'db/resCoverPoster/manual-{i+1:02d}.jpg',i,title)
        scenic(RUN/f'media/atlas/{i+1:02d}.jpg',i,title,(960,640))
        scenic(RUN/f'media/comic/{i+1:02d}.jpg',i,f'山海游记 · 第{i+1}页',(640,900))
        im=Image.new('RGB',(240,300),[(40,96,111),(105,87,130),(138,101,73),(63,116,91)][i%4]);d=ImageDraw.Draw(im)
        d.ellipse((70,45,170,145),fill=(236,218,191));d.rounded_rectangle((35,160,205,330),radius=70,fill=(186,204,204))
        d.text((15,12),f'虚构人物 {i+1:02d}',font=ImageFont.truetype(FONT,18),fill='white'); im.save(RUN/f'db/performerFace/manual-{i+1:02d}.jpg',quality=83,optimize=True)
    # A small original PCM tone for audio import, with a real playable payload.
    audio=RUN/'media/manual-tone.wav'
    with wave.open(str(audio),'wb') as f:
        f.setparams((1,2,16000,0,'NONE','not compressed'))
        f.writeframes(b''.join(struct.pack('<h',int(1600*math.sin(2*math.pi*440*t/16000))) for t in range(16000*4)))
    (RUN/'media/README.txt').write_text('全部为手册脚本原创生成的虚构示例；可删除并重建。\n',encoding='utf-8')
    (RUN/'asset-manifest.json').write_text(json.dumps({'generator':'prepare_demo.py','seed':3000,'source':'Original procedural illustration; no external downloads','titles':TITLES},ensure_ascii=False,indent=2),encoding='utf-8')
    print(RUN)
if __name__=='__main__': main()
