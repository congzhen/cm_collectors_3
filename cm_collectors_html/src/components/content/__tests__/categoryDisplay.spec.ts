import { expect, it } from 'vitest';
import { categoryCardDisplay, categoryGridDisplay } from '../categoryDisplay';
import type { I_resource } from '@/dataType/resource.dataType';

const cover = { coverPosterWidth: 900, coverPosterHeight: 1600 } as I_resource;
it('分类封面遵循普通封面的宽高锁定规则', () => {
  expect(categoryCardDisplay({ coverPosterHeightStatus: true, coverPosterHeightBase: 320 }, cover))
    .toMatchObject({ '--category-cover-width': '180px', '--category-cover-ratio': '180 / 320' });
  expect(categoryCardDisplay({ coverPosterWidthStatus: true, coverPosterWidthBase: 240 }, cover))
    .toMatchObject({ '--category-cover-width': '240px', '--category-cover-ratio': '240 / 1600' });
  expect(categoryCardDisplay({ coverPosterWidthStatus: true, coverPosterWidthBase: 240, coverPosterHeightStatus: true, coverPosterHeightBase: 300 }, cover))
    .toMatchObject({ '--category-cover-width': '240px', '--category-cover-ratio': '240 / 300' });
  expect(categoryCardDisplay({}, cover)).toMatchObject({ '--category-cover-ratio': '900 / 1600' });
});
it('缺少图片尺寸时使用安全尺寸，字号、填充和标题对齐可配置', () => {
  expect(categoryCardDisplay({ coverDisplayTagFontSize: 20, coverImageFit: 'contain', coverTitleAlign: 'right' }))
    .toMatchObject({ '--category-cover-width': '162px', '--category-cover-ratio': '162 / 210', '--category-badge-size': '20px', '--category-cover-fit': 'contain', '--category-caption-align': 'flex-end' });
});
it('资源间距支持零值，左右空距与对齐沿用快捷设置', () => {
  expect(categoryGridDisplay({ coverPosterGap: 0, contentPadding: 10, resourceJustifyContent: 'space-between' }))
    .toMatchObject({ gap: '0px', paddingInline: '10%', justifyContent: 'space-between' });
});
