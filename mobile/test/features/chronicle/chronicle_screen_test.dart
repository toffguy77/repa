import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:repa/core/api/api_service.dart';
import 'package:repa/core/providers/api_provider.dart';
import 'package:repa/core/theme/app_theme.dart';
import 'package:repa/features/chronicle/presentation/chronicle_screen.dart';

class MockDio extends Mock implements Dio {}

Future<void> pumpScreen(WidgetTester tester, Map<String, dynamic> data) async {
  final mockDio = MockDio();
  when(() => mockDio.get('/groups/g1/chronicle')).thenAnswer((_) async => Response(
        data: {'data': data},
        statusCode: 200,
        requestOptions: RequestOptions(path: '/groups/g1/chronicle'),
      ));

  await tester.pumpWidget(ProviderScope(
    overrides: [apiServiceProvider.overrideWithValue(ApiService(mockDio))],
    child: MaterialApp(
      theme: AppTheme.dark,
      home: const ChronicleScreen(groupId: 'g1'),
    ),
  ));
  await tester.pumpAndSettle();
}

void main() {
  testWidgets('a new group sees what will fill its history', (tester) async {
    await pumpScreen(tester, {
      'seasons': <dynamic>[],
      'standing': {'available': false, 'reason': 'рано', 'members': <dynamic>[]},
      'seasons_total': 0,
      'seasons_shown': 0,
    });

    expect(find.text('Пока ничего не произошло'), findsOneWidget);
    // Names what will fill it, rather than only reporting that it is empty.
    expect(find.textContaining('После первой репы'), findsOneWidget);
  });

  testWidgets('a group with history sees its seasons newest first',
      (tester) async {
    await pumpScreen(tester, {
      'seasons': [
        {
          'season_id': 's2',
          'number': 2,
          'reveal_at': '2026-09-25T17:00:00Z',
          'entries': [
            {
              'question_id': 'q1',
              'question_text': 'Кто чаще всех смеётся?',
              'category': 'FUNNY',
              'user_id': 'u1',
              'username': 'alice',
              'percentage': 75.0,
              'vote_count': 9,
              'total_voters': 12,
              'small_sample': false,
            }
          ],
        },
        {
          'season_id': 's1',
          'number': 1,
          'reveal_at': '2026-09-18T17:00:00Z',
          'entries': <dynamic>[],
        },
      ],
      'standing': {
        'available': true,
        'members': [
          {
            'user_id': 'u1',
            'username': 'alice',
            'rank': 1,
            'ranked': true,
            'seasons_played': 2,
          }
        ],
      },
      'seasons_total': 2,
      'seasons_shown': 2,
    });

    expect(find.text('Репа №2'), findsOneWidget);
    expect(find.text('Репа №1'), findsOneWidget);
    expect(find.text('alice'), findsNWidgets(2)); // once in the entry, once in the standing
    // The percentage belongs to the chronicle entry only. The standing shows a place and the history
    // it rests on, never an accuracy figure — that is the whole point of withholding it.
    expect(find.text('75%'), findsOneWidget);
    expect(find.text('за 2 репы'), findsOneWidget);
    expect(find.text('Пока ничего не произошло'), findsNothing);
  });

  testWidgets('the cap is stated rather than hidden', (tester) async {
    await pumpScreen(tester, {
      'seasons': [
        {
          'season_id': 's50',
          'number': 50,
          'reveal_at': '2026-09-25T17:00:00Z',
          'entries': <dynamic>[],
        }
      ],
      'standing': {'available': false, 'reason': '', 'members': <dynamic>[]},
      'seasons_total': 50,
      'seasons_shown': 1,
    });

    expect(find.textContaining('Показаны последние 1 из 50'), findsOneWidget);
  });
}
