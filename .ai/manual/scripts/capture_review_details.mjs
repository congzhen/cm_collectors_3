// Supplementary documentation views. Opens real forms and runs path SEARCH
// only; never submits a replacement, migration, deletion or batch edit.
import fs from 'node:fs';
import path from 'node:path';
import {api,browser,pageReady,shot,run,state} from './browser.mjs';
const b=await browser();const p=await b.newPage({viewport:{width:1440,height:1000},deviceScaleFactor:1});
p.setDefaultTimeout(15000);
const receipt={screenshots:[],write_actions:[],completed:false};
try{
 await p.goto('http://127.0.0.1:18346/setting',{waitUntil:'networkidle'});
 await p.getByRole('tab',{name:'文件数据库',exact:true}).click();await pageReady(p);
 await p.getByRole('button',{name:'路径替换',exact:true}).click();
 const dialog=p.getByRole('dialog',{name:'资源路径替换器',exact:true});
 const inputs=dialog.locator('.condition-item input.el-input__inner');
 const r=await api('/resource/info/'+state.resources[0].id);
 const directory=path.dirname(r.dramaSeries[0].src);
 await dialog.locator('.el-select').click();
 await p.getByRole('option',{name:'山海影像',exact:true}).click();
 await dialog.getByText('检索路径:',{exact:true}).click();
 await inputs.nth(0).fill(path.join(path.basename(path.dirname(directory)),path.basename(directory)));
 await dialog.getByRole('button',{name:'检索',exact:true}).click();await pageReady(p);
 const matches=await dialog.locator('.result-list li').count();
 if(matches<12)throw Error('Expected a full page of real generated path matches');
 await shot(p,'library-paths');receipt.screenshots.push({name:'library-paths',real_matches:matches});
 await p.goto('http://127.0.0.1:18346/performer/'+state.perBase,{waitUntil:'networkidle'});await pageReady(p);
 await p.getByRole('slider').last().focus();await p.keyboard.press('End');
 await p.getByRole('button',{name:'批量管理',exact:true}).click();await pageReady(p);
 const cards=p.locator('.performer-batch-card');
 if(await cards.count()<12)throw Error('Expected populated performer list');
 for(let i=0;i<3;i++)await cards.nth(i).click();
 await shot(p,'performer-batch');receipt.screenshots.push({name:'performer-batch',selected:3});
 receipt.completed=true;
}finally{
 await b.close();receipt.date=new Date().toISOString();fs.writeFileSync(path.join(run,'review-details-verification.json'),JSON.stringify(receipt,null,2));
}
console.log(JSON.stringify(receipt));
