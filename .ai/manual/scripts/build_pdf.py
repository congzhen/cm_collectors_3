"""Build the editable manual, links, bookmarks, source map and layout audit.
Reads only explicit manual sources and mapped application SOURCE files.
"""
from pathlib import Path
import argparse, hashlib, json, math, re
from xml.sax.saxutils import escape
from io import BytesIO
from PIL import Image, ImageOps, ImageDraw, ImageFont
from reportlab.pdfgen.canvas import Canvas
from reportlab.pdfbase import pdfmetrics
from reportlab.pdfbase.ttfonts import TTFont
from reportlab.platypus import Paragraph, Table, TableStyle
from reportlab.lib.styles import ParagraphStyle
from reportlab.lib.colors import HexColor, Color, white
from reportlab.lib.utils import ImageReader
from pypdf import PdfReader

BASE=Path(__file__).resolve().parents[1]
ROOT=BASE.parents[1]
OUT=ROOT/'CM Collectors 3 使用说明.pdf'
LOCAL=BASE/'.local';LOCAL.mkdir(exist_ok=True)
W,H=595.276,841.89; M=42; CW=W-2*M
INK=HexColor('#20353e');TEAL=HexColor('#087f86');MUTED=HexColor('#60747c');PALE=HexColor('#eef5f5')
FONT='C:/Windows/Fonts/msyh.ttc'; BOLD='C:/Windows/Fonts/msyhbd.ttc'
pdfmetrics.registerFont(TTFont('CN',FONT));pdfmetrics.registerFont(TTFont('CNB',BOLD))
STYLES={
 'lead':ParagraphStyle('lead',fontName='CN',fontSize=10.5,leading=17,textColor=MUTED,wordWrap='CJK'),
 'body':ParagraphStyle('body',fontName='CN',fontSize=10.3,leading=17,textColor=INK,wordWrap='CJK'),
 'small':ParagraphStyle('small',fontName='CN',fontSize=8.2,leading=12,textColor=MUTED,wordWrap='CJK'),
 'cell':ParagraphStyle('cell',fontName='CN',fontSize=9.2,leading=15,textColor=INK,wordWrap='CJK'),
}
audit=[];page_map={}
def paragraph(c,text,x,top,width,style='body'):
 p=Paragraph(escape(text).replace('\n','<br/>'),STYLES[style]);_,height=p.wrap(width,H)
 p.drawOn(c,x,H-top-height)
 return top+height
def txt(c,s,x,top,size=10,font='CN',color=INK):
 if x+pdfmetrics.stringWidth(s,font,size)>W-M+0.1:raise ValueError('text exceeds right margin: '+s)
 c.setFont(font,size);c.setFillColor(color);c.drawString(x,H-top-size*.82,s)
def header(c,chapter,number):
 txt(c,'CM COLLECTORS 3',M,23,8,'CNB',TEAL);c.setStrokeColor(HexColor('#dce6e7'));c.line(M,H-42,W-M,H-42)
 c.setFont('CN',8);c.setFillColor(MUTED);c.drawRightString(W-M,H-30,chapter)
 c.line(M,38,W-M,38);txt(c,'使用说明 · 2026-10-26',M,H-29,7.5,color=MUTED)
 c.setFont('CN',8);c.drawRightString(W-M,23,f'{number:02d}')
def image_block(c,p,top):
 name=p['image'];file=BASE/'assets/screenshots'/f'{name}.jpg'
 if not file.exists():raise FileNotFoundError(file)
 im=Image.open(file)
 if p.get('crop'):im=im.crop(tuple(p['crop']))
 maxh=337 if not p.get('portrait') else 410
 maxw=CW if not p.get('portrait') else 190
 scale=min(maxw/im.width,maxh/im.height)
 iw,ih=im.width*scale,im.height*scale
 ix=M+(CW-iw)/2 if not p.get('portrait') else M
 # Non-destructive placement: the screenshot remains the original tracked asset.
 c.setFillColor(PALE);c.roundRect(ix-3,H-top-ih-3,iw+6,ih+6,4,stroke=0,fill=1)
 if p.get('crop'):
  buffer=BytesIO();im.save(buffer,format='JPEG',quality=90,optimize=True);buffer.seek(0)
  picture=ImageReader(buffer)
 else:picture=str(file)
 c.drawImage(picture,ix,H-top-ih,width=iw,height=ih)
 caption=f"图 {p['_figure']} · {p.get('caption',p['title'])}"
 caption_end=paragraph(c,caption,ix,top+ih+8,iw,'small')
 return caption_end+10
def blocks(c,p,top,x=M,width=CW):
 for i,b in enumerate(p['blocks'],1):
  txt(c,f'{i:02d}',x,top+1,9,'CNB',TEAL)
  txt(c,b['title'],x+24,top,11,'CNB')
  top=paragraph(c,b['text'],x+24,top+23,width-24)+13
 return top
def rows(c,data,top):
 cells=[[Paragraph(escape(str(s)),STYLES['cell'])for s in row]for row in data]
 t=Table(cells,colWidths=[CW*.27,CW*.73],hAlign='LEFT')
 t.setStyle(TableStyle([('BACKGROUND',(0,0),(-1,0),HexColor('#dceeed')),('VALIGN',(0,0),(-1,-1),'TOP'),('ROWBACKGROUNDS',(0,1),(-1,-1),[HexColor('#f3f7f7'),white]),('LINEBELOW',(0,0),(-1,-1),.4,HexColor('#d8e4e5')),('LEFTPADDING',(0,0),(-1,-1),10),('RIGHTPADDING',(0,0),(-1,-1),10),('TOPPADDING',(0,0),(-1,-1),9),('BOTTOMPADDING',(0,0),(-1,-1),9)]))
 _,height=t.wrap(CW,H);t.drawOn(c,M,H-top-height);return top+height

def main():
 manifest=json.loads((BASE/'manual.json').read_text(encoding='utf-8'))
 chapters=[json.loads((BASE/'chapters'/f'{name}.json').read_text(encoding='utf-8'))for name in manifest['chapters']]
 chapter_numbers={ch['id']:i+1 for i,ch in enumerate(chapters)}
 ids=[p['id']for ch in chapters for p in ch['pages']]
 if len(ids)!=len(set(ids)):raise ValueError('duplicate stable page IDs')
 entries=[]
 for ch in chapters:
  entries.append(('chapter',ch['title'],ch['id']))
  entries.extend(('page',p['title'],p['id'])for p in ch['pages'])
 toc_chunks=[entries[i:i+27]for i in range(0,len(entries),27)]
 first=2+len(toc_chunks)
 for n,p in enumerate(p for ch in chapters for p in ch['pages']):page_map[p['id']]=first+n
 for ch in chapters:page_map[ch['id']]=page_map[ch['pages'][0]['id']]
 c=Canvas(str(OUT),pagesize=(W,H),pageCompression=1)
 c.setTitle(manifest['title']);c.setAuthor('CM Collectors 3');c.setSubject('资源管理、显示形式、标签、播放与设置的图文使用手册')
 c.setFillColor(HexColor('#10343e'));c.rect(0,0,W,H,stroke=0,fill=1)
 c.setFillColor(HexColor('#1c5360'))
 c.circle(540,655,220,stroke=0,fill=1);c.setFillColor(HexColor('#276a70'));c.circle(570,665,145,stroke=0,fill=1)
 txt(c,'CM / COLLECTORS 3',M,82,13,'CNB',HexColor('#a2d6d4'))
 txt(c,'把收藏整理好',M,175,39,'CNB',white);txt(c,'让每次浏览都有方向',M,234,25,'CN',white)
 txt(c,'使用说明',M,337,46,'CNB',white)
 txt(c,'资源管理 · 13 种显示形式 · 标签体系 · 播放与进阶工具',M,410,11,'CN',HexColor('#c6dcdf'))
 txt(c,'2026.10.26  /  第 1.0 版',M,687,12,'CNB',white)
 txt(c,'CM Collectors 3 · 图文使用说明',M,726,10,'CN',HexColor('#c6dcdf'))
 c.showPage()
 number=2
 for ix,chunk in enumerate(toc_chunks):
  header(c,'目录',number);txt(c,'按功能查阅',M,65,26,'CNB');txt(c,f'目录 {ix+1} / {len(toc_chunks)} · 点击条目跳转',M,108,10,color=MUTED)
  top=155
  for kind,title,id in chunk:
   isch=kind=='chapter';font='CNB'if isch else'CN';color=TEAL if isch else INK
   label=(f'{chapter_numbers[id]:02d}  ' if isch else '    ')+title
   txt(c,label,M,top,11 if isch else 10,font,color)
   c.setFont('CN',9);c.setFillColor(MUTED);c.drawRightString(W-M,H-top-9,str(page_map[id]))
   c.linkRect('',id,(M,H-top-20,W-M,H-top+3),relative=0,thickness=0)
   top+=22
  c.showPage();number+=1
 for ch in chapters:
  source_hashes={}
  for source in ch['sources']:
   file=ROOT/source
   if not file.is_file():raise FileNotFoundError(file)
   if file.suffix.lower() in ['.db','.jpg','.png','.yaml']:raise ValueError('application runtime data may not be read')
   source_hashes[source]=hashlib.sha256(file.read_bytes()).hexdigest()
  figure_number=0
  for pi,p in enumerate(ch['pages']):
   if p.get('image'):
    figure_number+=1;p['_figure']=f"{chapter_numbers[ch['id']]:02d}-{figure_number:02d}"
   header(c,ch['title'],number)
   if pi==0:c.bookmarkPage(ch['id']);c.addOutlineEntry(ch['title'],ch['id'],level=0)
   c.bookmarkPage(p['id']);c.addOutlineEntry(p['title'],p['id'],level=1)
   txt(c,f"第 {chapter_numbers[ch['id']]:02d} 章 · {pi+1:02d} / {len(ch['pages']):02d}",M,61,8,'CNB',TEAL)
   txt(c,p['title'],M,82,23,'CNB')
   top=paragraph(c,p['lead'],M,121,CW,'lead')+18
   if p.get('portrait'):
    end=image_block(c,p,top);body_end=blocks(c,p,top,M+216,CW-216);top=max(end,body_end)
   else:
    if p.get('image'):top=image_block(c,p,top)
    top=blocks(c,p,top)
   if p.get('rows'):top=rows(c,p['rows'],top+8)
   if top>H-58:raise ValueError(f"page overflow {p['id']}: bottom={top:.1f}, limit={H-58:.1f}")
   audit.append({'id':p['id'],'chapter':ch['id'],'title':p['title'],'pdf_page':number,'bottom':round(top,1),'image':p.get('image'),'source_hashes':source_hashes,'verification':p.get('verification')})
   c.showPage();number+=1
 c.save()
 pdf=PdfReader(str(OUT))
 if len(pdf.pages)!=number-1:raise ValueError('page count mismatch')
 for entry in audit:
  text=pdf.pages[entry['pdf_page']-1].extract_text()
  if entry['title']not in text:raise ValueError('missing Chinese text: '+entry['id'])
 (BASE/'feature-map.json').write_text(json.dumps({'edition':manifest['edition'],'date':manifest['date'],'output':OUT.name,'pages':len(pdf.pages),'features':audit},ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
 (LOCAL/'layout-audit.json').write_text(json.dumps({'pages':len(pdf.pages),'size_bytes':OUT.stat().st_size,'overflow_count':0,'entries':audit},ensure_ascii=False,indent=2),encoding='utf-8')
 print(json.dumps({'pdf':str(OUT),'pages':len(pdf.pages),'size_mb':round(OUT.stat().st_size/1024**2,2),'overflow':0},ensure_ascii=False))
if __name__=='__main__':main()
