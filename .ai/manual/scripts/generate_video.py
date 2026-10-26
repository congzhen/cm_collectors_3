from pathlib import Path
import subprocess, sys, shutil
BASE=Path(__file__).resolve().parents[1]
RUN=BASE/'.local/runtime'
sys.path.insert(0,str(BASE/'.local/python'))
import imageio_ffmpeg
tool=Path(imageio_ffmpeg.get_ffmpeg_exe())
(RUN/'ffmpeg').mkdir(exist_ok=True)
shutil.copy2(tool,RUN/'ffmpeg/ffmpeg.exe')
for i in range(24):
    out=RUN/f'media/videos/CM-{i+1:03d}.mp4'
    if out.exists(): continue
    subprocess.run([str(tool),'-hide_banner','-loglevel','error','-y','-loop','1','-framerate','10','-i',str(RUN/f'media/atlas/{i+1:02d}.jpg'),'-f','lavfi','-i',f'sine=frequency={220+i*12}:sample_rate=22050','-t','32','-vf','scale=640:426,format=yuv420p','-c:v','libx264','-preset','ultrafast','-crf','32','-c:a','aac','-b:a','32k','-movflags','+faststart',str(out)],check=True)
print('24 original playable videos ready')
