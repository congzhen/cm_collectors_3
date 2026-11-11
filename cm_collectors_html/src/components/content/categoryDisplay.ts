import type { CSSProperties } from 'vue';
import type { I_config_app } from '@/dataType/config.dataType';
import type { I_resource } from '@/dataType/resource.dataType';

const positive = (value: number | undefined, fallback: number) =>
  Number.isFinite(value) && value! > 0 ? value! : fallback;

export function categoryCardDisplay(config: Partial<I_config_app>, resource?: I_resource): CSSProperties {
  const originalWidth = positive(resource?.coverPosterWidth, 162);
  const originalHeight = positive(resource?.coverPosterHeight, 210);
  let width = originalWidth, height = originalHeight;
  if (config.coverPosterWidthStatus) width = positive(config.coverPosterWidthBase, 316);
  if (config.coverPosterHeightStatus) {
    height = positive(config.coverPosterHeightBase, 218);
    if (!config.coverPosterWidthStatus) width = height / originalHeight * originalWidth;
  }
  const align = ['left', 'center', 'right'].includes(config.coverTitleAlign || '') ? config.coverTitleAlign! : 'left';
  return {
    '--category-cover-width': `${width}px`,
    '--category-cover-ratio': `${width} / ${height}`,
    '--category-badge-size': `${positive(config.coverDisplayTagFontSize, 12)}px`,
    '--category-title-align': align,
    '--category-caption-align': align === 'center' ? 'center' : align === 'right' ? 'flex-end' : 'flex-start',
    '--category-cover-fit': config.coverImageFit || 'cover',
  };
}

export function categoryGridDisplay(config: Partial<I_config_app>): CSSProperties {
  return {
    gap: `${Math.max(0, config.coverPosterGap ?? 4.8)}px`,
    paddingInline: `${Math.max(0, Math.min(50, config.contentPadding ?? 0))}%`,
    justifyContent: config.resourceJustifyContent || 'flex-start',
    alignItems: 'flex-start',
  };
}
