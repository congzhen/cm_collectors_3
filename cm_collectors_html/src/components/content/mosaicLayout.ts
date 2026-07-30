export interface I_mosaicSource<T> {
  data: T;
  aspectRatio: number;
  preferredSpan: 1 | 2;
}

export interface I_mosaicLayoutItem<T> extends I_mosaicSource<T> {
  span: number;
  width: number;
}

export interface I_mosaicLayoutRow<T> {
  height: number;
  items: I_mosaicLayoutItem<T>[];
}

type I_pendingMosaicItem<T> = I_mosaicSource<T>;

type T_pendingMosaicRow<T> = Array<
  I_pendingMosaicItem<T> & { span: number }
>;

const clamp = (value: number, minimum: number, maximum: number) => {
  return Math.min(maximum, Math.max(minimum, value));
}

const getDesiredSpan = (remainingColumns: number) => {
  if (remainingColumns === 1) return 1;
  return remainingColumns % 2 === 0 ? 2 : 1;
}

const takeBestItem = <T>(
  pendingItems: I_pendingMosaicItem<T>[],
  remainingColumns: number,
  reorderWindow: number,
) => {
  const searchLength = Math.min(reorderWindow, pendingItems.length);
  const desiredSpan = getDesiredSpan(remainingColumns);
  let selectedIndex = -1;

  for (let index = 0; index < searchLength; index++) {
    if (pendingItems[index].preferredSpan === desiredSpan) {
      selectedIndex = index;
      break;
    }
  }

  if (selectedIndex < 0) {
    for (let index = 0; index < searchLength; index++) {
      if (pendingItems[index].preferredSpan <= remainingColumns) {
        selectedIndex = index;
        break;
      }
    }
  }

  if (selectedIndex < 0) selectedIndex = 0;
  return pendingItems.splice(selectedIndex, 1)[0];
}

const calculateRowHeight = <T>(
  items: I_mosaicLayoutItem<T>[],
  unitWidth: number,
) => {
  const idealHeights = items
    .map((item) => item.width / item.aspectRatio)
    .filter((height) => Number.isFinite(height) && height > 0);

  if (idealHeights.length === 0) return Math.round(unitWidth * 1.25);

  const logarithmicAverage =
    idealHeights.reduce((sum, height) => sum + Math.log(height), 0) /
    idealHeights.length;
  const preferredHeight = Math.exp(logarithmicAverage);

  return Math.round(
    clamp(preferredHeight, unitWidth * 0.8, unitWidth * 1.65),
  );
}

export const createMosaicRows = <T>(
  sources: I_mosaicSource<T>[],
  columnCount: number,
  containerWidth: number,
  gap: number,
  reorderWindow = 16,
): I_mosaicLayoutRow<T>[] => {
  if (sources.length === 0 || containerWidth <= 0) return [];

  const normalizedColumnCount = Math.max(2, Math.floor(columnCount));
  const normalizedGap = Math.max(0, gap);
  const normalizedWindow = Math.max(1, Math.floor(reorderWindow));
  const unitWidth =
    (containerWidth - (normalizedColumnCount - 1) * normalizedGap) /
    normalizedColumnCount;

  if (unitWidth <= 0) return [];

  const pendingItems: I_pendingMosaicItem<T>[] = sources.map(
    (source) => ({
      ...source,
      aspectRatio:
        Number.isFinite(source.aspectRatio) && source.aspectRatio > 0
          ? source.aspectRatio
          : 2 / 3,
    }),
  );
  const pendingRows: T_pendingMosaicRow<T>[] = [];

  while (pendingItems.length > 0) {
    let remainingColumns = normalizedColumnCount;
    const selectedItems: T_pendingMosaicRow<T> = [];

    while (remainingColumns > 0 && pendingItems.length > 0) {
      const selectedItem = takeBestItem(
        pendingItems,
        remainingColumns,
        normalizedWindow,
      );
      const span = Math.min(selectedItem.preferredSpan, remainingColumns);
      selectedItems.push({ ...selectedItem, span });
      remainingColumns -= span;
    }

    pendingRows.push(selectedItems);
    if (pendingItems.length === 0) break;
    if (remainingColumns !== 0) break;
  }

  return pendingRows.map((rowItems) => {
    const items = rowItems.map<I_mosaicLayoutItem<T>>((item) => ({
      data: item.data,
      aspectRatio: item.aspectRatio,
      preferredSpan: item.preferredSpan,
      span: item.span,
      width:
        item.span * unitWidth + (item.span - 1) * normalizedGap,
    }));

    return {
      height: calculateRowHeight(items, unitWidth),
      items,
    };
  });
}
