import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:repa/core/theme/app_tokens.dart';

/// Relative luminance per WCAG 2.1.
double _luminance(Color c) {
  double channel(double v) =>
      v <= 0.03928 ? v / 12.92 : math.pow((v + 0.055) / 1.055, 2.4).toDouble();
  return 0.2126 * channel(c.r) + 0.7152 * channel(c.g) + 0.0722 * channel(c.b);
}

double contrastRatio(Color a, Color b) {
  final la = _luminance(a);
  final lb = _luminance(b);
  final hi = math.max(la, lb);
  final lo = math.min(la, lb);
  return (hi + 0.05) / (lo + 0.05);
}

/// A pairing the app actually renders, and the ratio it must clear.
class _Pair {
  const _Pair(this.fg, this.bg, this.min);
  final String fg;
  final String bg;
  final double min;
}

/// 4.5:1 for text; 3:1 for large text and for the boundary of interactive controls.
const _surfaces = ['canvas', 'surface', 'surfaceRaised'];

List<_Pair> _pairs() => [
      // Body and secondary text on every surface they can land on.
      for (final s in _surfaces) ...[
        _Pair('textPrimary', s, 4.5),
        _Pair('textSecondary', s, 4.5),
      ],
      // Status and brand colours are used for icons, emphasis and large text.
      for (final s in _surfaces) ...[
        _Pair('accent', s, 3.0),
        _Pair('energy', s, 3.0),
        _Pair('success', s, 3.0),
        _Pair('warning', s, 3.0),
        _Pair('danger', s, 3.0),
      ],
      // Filled controls carry text on top of them.
      const _Pair('onAccentFill', 'accentFill', 4.5),
      const _Pair('onAccentFill', 'accentFillPressed', 4.5),
      const _Pair('onEnergyFill', 'energyFill', 4.5),
    ];

void main() {
  group('WCAG AA contrast', () {
    for (final entry in {
      'dark': AppColorTokens.dark,
      'light': AppColorTokens.light,
    }.entries) {
      group(entry.key, () {
        final colors = entry.value.all;

        for (final pair in _pairs()) {
          test('${pair.fg} on ${pair.bg} clears ${pair.min}:1', () {
            final fg = colors[pair.fg];
            final bg = colors[pair.bg];
            expect(fg, isNotNull, reason: 'unknown token ${pair.fg}');
            expect(bg, isNotNull, reason: 'unknown token ${pair.bg}');

            final ratio = contrastRatio(fg!, bg!);
            expect(
              ratio,
              greaterThanOrEqualTo(pair.min),
              reason: '${pair.fg} on ${pair.bg} in ${entry.key} is '
                  '${ratio.toStringAsFixed(2)}:1, needs ${pair.min}:1',
            );
          });
        }
      });
    }

    test('the border token is visible against its surfaces', () {
      for (final palette in [AppColorTokens.dark, AppColorTokens.light]) {
        for (final surface in [palette.surface, palette.surfaceRaised]) {
          expect(
            contrastRatio(palette.border, surface),
            greaterThan(1.05),
            reason: 'a hairline that matches its surface is not a hairline',
          );
        }
      }
    });
  });

  group('contrastRatio helper', () {
    test('black on white is 21:1', () {
      expect(
        contrastRatio(const Color(0xFF000000), const Color(0xFFFFFFFF)),
        closeTo(21, 0.01),
      );
    });

    test('is symmetric', () {
      const a = Color(0xFF9B6DFF);
      const b = Color(0xFF0B0712);
      expect(contrastRatio(a, b), closeTo(contrastRatio(b, a), 1e-9));
    });

    test('identical colours are 1:1', () {
      const c = Color(0xFF123456);
      expect(contrastRatio(c, c), closeTo(1, 1e-9));
    });
  });
}
