import { describe, expect, it } from 'vitest';

import { createMosaicRows, type I_mosaicSource } from '../mosaicLayout';

interface I_testResource {
  id: string;
}

const source = (
  id: string,
  aspectRatio: number,
  preferredSpan: 1 | 2,
): I_mosaicSource<I_testResource> => ({
  data: { id },
  aspectRatio,
  preferredSpan,
});

describe('createMosaicRows', () => {
  it('fills complete rows and allows a trailing hole in the final row', () => {
    const rows = createMosaicRows(
      [
        source('landscape-1', 16 / 9, 2),
        source('portrait-1', 2 / 3, 1),
        source('landscape-2', 1.8, 2),
        source('portrait-2', 0.7, 1),
        source('portrait-3', 0.65, 1),
      ],
      4,
      800,
      10,
    );

    expect(rows.length).toBeGreaterThan(0);
    for (const row of rows.slice(0, -1)) {
      expect(row.items.reduce((sum, item) => sum + item.span, 0)).toBe(4);
      expect(
        row.items.reduce((sum, item) => sum + item.width, 0) +
          (row.items.length - 1) * 10,
      ).toBeCloseTo(800);
    }
    expect(
      rows.at(-1)?.items.reduce((sum, item) => sum + item.span, 0),
    ).toBeLessThanOrEqual(4);
  });

  it('uses a nearby portrait to complete an odd-width row', () => {
    const rows = createMosaicRows(
      [
        source('landscape-1', 16 / 9, 2),
        source('landscape-2', 16 / 9, 2),
        source('portrait', 2 / 3, 1),
      ],
      3,
      600,
      8,
    );

    expect(rows[0].items.map((item) => item.data.id)).toEqual(
      expect.arrayContaining(['portrait']),
    );
    expect(
      rows[0].items.some((item) => item.data.id.startsWith('landscape-')),
    ).toBe(true);
    expect(rows[0].items.map((item) => item.span)).toEqual([1, 2]);
  });

  it('keeps the final row deterministic without expanding its covers', () => {
    const sources = [
      source('portrait', 2 / 3, 1),
      source('landscape', 16 / 9, 2),
    ];
    const first = createMosaicRows(sources, 6, 900, 10);
    const second = createMosaicRows(sources, 6, 900, 10);

    expect(first).toEqual(second);
    expect(first[0].items.reduce((sum, item) => sum + item.span, 0)).toBe(3);
    expect(first[0].items.map((item) => item.span)).toEqual([2, 1]);
  });

  it('keeps every resource with different column counts and aspect mixes', () => {
    for (let columnCount = 2; columnCount <= 20; columnCount++) {
      for (let itemCount = 1; itemCount <= 40; itemCount++) {
        for (const landscapeInterval of [1, 2, 3, 5]) {
          const sources = Array.from({ length: itemCount }, (_, index) =>
            index % landscapeInterval === 0
              ? source(`landscape-${index}`, 16 / 9, 2)
              : source(`portrait-${index}`, 2 / 3, 1),
          );
          const rows = createMosaicRows(
            sources,
            columnCount,
            1200,
            10,
          );
          const items = rows.flatMap((row) => row.items);

          expect(items).toHaveLength(itemCount);
          expect(new Set(items.map((item) => item.data.id)).size).toBe(
            itemCount,
          );
          for (const row of rows.slice(0, -1)) {
            expect(
              row.items.reduce((sum, item) => sum + item.span, 0),
            ).toBe(columnCount);
          }
          expect(
            rows
              .at(-1)
              ?.items.reduce((sum, item) => sum + item.span, 0),
          ).toBeLessThanOrEqual(columnCount);
        }
      }
    }
  });
});
