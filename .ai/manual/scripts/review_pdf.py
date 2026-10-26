"""Render every page for human review; create numbered contact sheets."""
from pathlib import Path
import subprocess,shutil,json,os,re
from PIL import Image,ImageOps,ImageDraw,ImageFont
from pypdf import PdfReader
B=Path(__file__).resolve().parents[1];R=B.parents[1]
pdf=R/'CM Collectors 3 使用说明.pdf';out=B/'.local/render';out.mkdir(parents=True,exist_ok=True)
bundle=Path(os.environ.get('USERPROFILE',str(Path.home())))/'.cache/codex-runtimes/codex-primary-runtime/dependencies/native/poppler/Library/bin/pdftoppm.exe'
tool=shutil.which('pdftoppm') or (str(bundle)if bundle.exists()else None)
if not tool:raise SystemExit('pdftoppm is required; locate the bundled runtime or install Poppler')
# Remove only this script's generated page/contact files; shorter editions must
# not retain pages from a previous, longer edition.
for path in out.iterdir():
 if path.is_file() and re.fullmatch(r'(page-\d+\.png|contact-\d+\.jpg)',path.name):path.unlink()
subprocess.run([tool,'-r','90','-png',str(pdf),str(out/'page')],check=True)
count=len(PdfReader(str(pdf)).pages)
pages=sorted(out.glob('page-*.png'))
if len(pages)!=count:raise SystemExit('render count mismatch; inspect stale render files')
font=ImageFont.truetype('C:/Windows/Fonts/arial.ttf',14)
for start in range(0,len(pages),8):
 sheet=Image.new('RGB',(1240,940),'#d8e0e3');d=ImageDraw.Draw(sheet)
 for i,path in enumerate(pages[start:start+8]):
  im=Image.open(path).convert('RGB');im.thumbnail((290,410))
  x=(i%4)*310+(310-im.width)//2;y=(i//4)*470+32
  sheet.paste(im,(x,y));d.text((x,y-22),f'PAGE {start+i+1:02d}',font=font,fill='#193841')
 sheet.save(out/f'contact-{start//8+1:02d}.jpg',quality=90)
print(json.dumps({'rendered_pages':count,'contact_sheets':(count+7)//8,'directory':str(out)}))
