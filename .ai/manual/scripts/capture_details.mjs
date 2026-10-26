import {api,browser,home,pageReady,shot,state} from './browser.mjs';
const b=await browser();const p=await b.newPage({viewport:{width:1440,height:1000},deviceScaleFactor:1});
p.setDefaultTimeout(12000);
const phase=process.argv[2]||'play';
try{
 if(phase==='play'){
  const r=state.resources[0];await p.goto('http://127.0.0.1:18346/play/movies/'+r.id,{waitUntil:'networkidle'});await pageReady(p);
  console.log((await p.locator('body').innerText()).slice(-2500));
  if(await p.locator('video').count()){await p.locator('video').first().evaluate(v=>{v.muted=true;v.currentTime=8;return v.play().catch(()=>{});});await p.waitForTimeout(1500);await p.locator('video').first().evaluate(v=>v.pause());}
  await shot(p,'player');
  if(await p.getByText('片段 / 预览',{exact:false}).count()){await p.getByText('片段 / 预览',{exact:false}).first().click({force:true});await shot(p,'player-moments');}
  await p.goto('http://127.0.0.1:18346/play/atlas/'+state.resources[24].id,{waitUntil:'networkidle'});await shot(p,'atlas');
  await p.goto('http://127.0.0.1:18346/play/comic/'+state.resources[48].id,{waitUntil:'networkidle'});await shot(p,'comic');
 }else if(phase==='advanced'){
  await p.goto('http://127.0.0.1:18346/setting',{waitUntil:'networkidle'});await p.getByRole('tab',{name:'视频转码与剪辑',exact:true}).click();await pageReady(p);await shot(p,'transcode-result');
  await p.getByRole('button',{name:'剪辑',exact:true}).first().click();await pageReady(p);await shot(p,'video-editor');
  await p.goto('http://127.0.0.1:18346/setting',{waitUntil:'networkidle'});await p.getByRole('tab',{name:'视频重复检测',exact:true}).click();await pageReady(p);
  console.log((await p.locator('body').innerText()).slice(-1800));
  await p.getByText('宽松',{exact:true}).click();
  const search=p.getByRole('button',{name:/查找|查询|检测/}).first();if(await search.count())await search.click();await shot(p,'duplicate-results');
  if(await p.getByRole('button',{name:/同屏复核/}).count()){await p.getByRole('button',{name:/同屏复核/}).first().click();await shot(p,'duplicate-compare');}
  await p.goto('http://127.0.0.1:18346/setting',{waitUntil:'networkidle'});await p.getByRole('tab',{name:'视频信息采集',exact:true}).click();await shot(p,'metadata-result');
 }else if(phase==='library'){
  await home(p);await p.getByRole('button',{name:'显示设置',exact:true}).click();await shot(p,'display-settings');
  await p.goto('http://127.0.0.1:18346/setting',{waitUntil:'networkidle'});await p.getByRole('tab',{name:'文件数据库',exact:true}).click();await p.getByRole('tab',{name:'山海影像',exact:true}).click();await pageReady(p);await p.getByRole('button',{name:/创建公共配置|打开公共配置/}).first().click();await shot(p,'shared-display');
  await home(p);await p.getByTitle('导入与刮削').click();await p.getByRole('button',{name:'添加文件夹位置',exact:true}).click();await pageReady(p);await shot(p,'file-manager');
 }else if(phase==='filters'){
  await home(p);const group=p.locator('.tag-logic-container').filter({hasText:'题材'});await group.locator('.tag-logic-text').hover();await p.waitForTimeout(700);
  console.log('visible menus',await p.locator('[role=menu]:visible').count());
  await p.screenshot({path:new URL('../assets/screenshots/tag-logic.jpg',import.meta.url).pathname.replace(/^\/([A-Z]:)/,'$1'),type:'jpeg',quality:86});
  await p.locator('.el-dropdown__popper:visible').getByText('多选(或)',{exact:true}).click();
  const side=p.locator('.tag-block').filter({has:p.locator('.tag-logic-container').filter({hasText:'题材'})});
  console.log('tagblock',await side.count());
  await p.getByText('自然',{exact:true}).first().click();await p.getByText('旅行',{exact:true}).first().click();await shot(p,'tag-filter');
  await p.getByTitle('搜索收藏',{exact:true}).hover();await p.waitForTimeout(700);await p.getByRole('button',{name:'保存当前搜索',exact:true}).click();await p.waitForTimeout(700);await p.getByTitle('搜索收藏',{exact:true}).hover();await shot(p,'search-favorite');
 }else if(phase==='moments'){
  await p.goto('http://127.0.0.1:18346/moments',{waitUntil:'networkidle'});await shot(p,'moments');
 }else if(phase==='backup'){
  console.log('backup result',JSON.stringify(await api('/database/autoBackup/run',{})));
  await p.goto('http://127.0.0.1:18346/setting',{waitUntil:'networkidle'});await p.getByRole('tab',{name:'自动备份',exact:true}).click();await shot(p,'setting-backup');
 }else if(phase==='files'){
  await home(p);await p.locator('.content-right-click-menu').first().click({button:'right'});await p.getByText('编辑',{exact:true}).click();await p.getByRole('button',{name:'选择资源',exact:true}).click();await pageReady(p);
  const dialog=p.getByRole('dialog',{name:'文件管理',exact:true});await dialog.locator('.file-name-block').first().click();await pageReady(p);await dialog.getByText('videos',{exact:true}).click();await shot(p,'file-manager');
 }
}catch(e){console.error(e.message);console.log((await p.locator('body').innerText()).slice(-2500));await shot(p,'debug-'+phase);process.exitCode=1;}
finally{await b.close();}
