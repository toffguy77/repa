import 'package:dio/dio.dart';
import 'package:firebase_analytics/firebase_analytics.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:repa/core/analytics/analytics_service.dart';
import 'package:repa/core/api/api_service.dart';
import 'package:repa/core/providers/api_provider.dart';
import 'package:repa/core/providers/auth_provider.dart';
import 'package:repa/core/theme/app_theme.dart';
import 'package:repa/features/auth/domain/user.dart';
import 'package:repa/features/groups/data/groups_repository.dart';
import 'package:repa/features/groups/domain/group.dart';
import 'package:repa/features/groups/presentation/create_group_screen.dart';

class MockDio extends Mock implements Dio {}

class _FakeFirebaseAnalytics extends Fake implements FirebaseAnalytics {}

/// Analytics is stubbed out because FirebaseAnalytics.instance needs a configured Firebase app that a
/// widget test has no way to provide — and the thing under test is the request body, not the event.
class FakeAnalytics extends AnalyticsService {
  FakeAnalytics() : super(_FakeFirebaseAnalytics());

  @override
  Future<void> logGroupCreated() async {}
}

/// An auth state with a chosen birth year, which is the only input to the switch's default position.
class FakeAuth extends AuthNotifier {
  FakeAuth(super.ref, int? birthYear) {
    state = AuthState(
      status: AuthStatus.authenticated,
      user: User(
        id: 'u1',
        username: 'tester',
        avatarEmoji: '\u{1F916}',
        birthYear: birthYear,
        createdAt: '2026-01-01T00:00:00Z',
      ),
    );
  }
}

Map<String, dynamic> createdGroupResponse() => {
      'data': {
        'group': {
          'id': 'g1',
          'name': 'Группа',
          'admin_id': 'u1',
          'invite_code': 'AB2CD3',
          'categories': ['FUNNY'],
          'kind_only': false,
          'created_at': '2026-01-01T00:00:00Z',
        },
        'invite_url': 'https://repa.app/join/AB2CD3',
      }
    };

/// Pumps the create-group screen and returns the bodies POSTed to /groups.
Future<List<Map<String, dynamic>>> pumpCreateScreen(
  WidgetTester tester, {
  required int? birthYear,
}) async {
  final mockDio = MockDio();
  final bodies = <Map<String, dynamic>>[];
  when(() => mockDio.post('/groups', data: any(named: 'data')))
      .thenAnswer((invocation) async {
    bodies.add(Map<String, dynamic>.from(
        invocation.namedArguments[const Symbol('data')] as Map));
    return Response(
      data: createdGroupResponse(),
      statusCode: 201,
      requestOptions: RequestOptions(path: '/groups'),
    );
  });
  // The screen refreshes the group list after creating, which would otherwise hit an unstubbed call.
  when(() => mockDio.get('/groups')).thenAnswer((_) async => Response(
        data: {
          'data': {'groups': <dynamic>[]}
        },
        statusCode: 200,
        requestOptions: RequestOptions(path: '/groups'),
      ));

  await tester.pumpWidget(ProviderScope(
    overrides: [
      apiServiceProvider.overrideWithValue(ApiService(mockDio)),
      authProvider.overrideWith((ref) => FakeAuth(ref, birthYear)),
      analyticsProvider.overrideWithValue(FakeAnalytics()),
    ],
    child: MaterialApp(theme: AppTheme.dark, home: const CreateGroupScreen()),
  ));
  await tester.pumpAndSettle();
  return bodies;
}

Future<void> fillAndSubmit(WidgetTester tester) async {
  await tester.enterText(find.byType(TextField).first, 'Наша группа');
  await tester.pump();
  await tester.tap(find.text('\u{1F602} Смешное'));
  await tester.pump();

  final button = find.widgetWithText(ElevatedButton, 'Создать');
  await tester.ensureVisible(button);
  await tester.pump();
  expect(tester.widget<ElevatedButton>(button).onPressed, isNotNull,
      reason: 'the form should be valid');
  await tester.tap(button);
  await tester.pump();
  await tester.pump(const Duration(milliseconds: 200));
}

void main() {
  group('the kind-only switch on group creation', () {
    testWidgets('defaults to on for a 15-year-old', (tester) async {
      await pumpCreateScreen(tester, birthYear: DateTime.now().year - 15);
      expect(tester.widget<Switch>(find.byType(Switch)).value, isTrue);
    });

    testWidgets('defaults to off for a 30-year-old', (tester) async {
      await pumpCreateScreen(tester, birthYear: DateTime.now().year - 30);
      expect(tester.widget<Switch>(find.byType(Switch)).value, isFalse);
    });

    testWidgets('defaults to on when the birth year is unknown',
        (tester) async {
      // Unknown age is treated as under 18, matching the server's default and the ROMANCE rule.
      await pumpCreateScreen(tester, birthYear: null);
      expect(tester.widget<Switch>(find.byType(Switch)).value, isTrue);
    });

    testWidgets('says what it does to the questions', (tester) async {
      await pumpCreateScreen(tester, birthYear: DateTime.now().year - 30);
      expect(find.text('Только добрые вопросы'), findsOneWidget);
      expect(find.textContaining('приятные и нейтральные вопросы'),
          findsOneWidget);
      expect(find.textContaining('Колкие вопросы'), findsOneWidget);
    });

    testWidgets('an untouched switch sends nothing', (tester) async {
      // The server owns the age rule. A client that always sent its own value would override it, and
      // would get it wrong whenever it does not know the creator's birth year.
      final bodies =
          await pumpCreateScreen(tester, birthYear: DateTime.now().year - 15);
      await fillAndSubmit(tester);

      expect(bodies, hasLength(1));
      expect(bodies.single.containsKey('kind_only'), isFalse,
          reason: 'an untouched switch must leave the default to the server');
    });

    testWidgets('a touched switch sends the chosen value', (tester) async {
      final bodies =
          await pumpCreateScreen(tester, birthYear: DateTime.now().year - 15);

      await tester.tap(find.byType(Switch));
      await tester.pumpAndSettle();
      expect(tester.widget<Switch>(find.byType(Switch)).value, isFalse);

      await fillAndSubmit(tester);
      expect(bodies.single['kind_only'], isFalse);
    });

    testWidgets('an adult can turn it on', (tester) async {
      final bodies =
          await pumpCreateScreen(tester, birthYear: DateTime.now().year - 30);

      await tester.tap(find.byType(Switch));
      await tester.pumpAndSettle();

      await fillAndSubmit(tester);
      expect(bodies.single['kind_only'], isTrue);
    });
  });

  group('GroupsRepository.createGroup', () {
    test('omits kind_only when the caller did not choose', () async {
      final mockDio = MockDio();
      Map<String, dynamic>? body;
      when(() => mockDio.post('/groups', data: any(named: 'data')))
          .thenAnswer((invocation) async {
        body = Map<String, dynamic>.from(
            invocation.namedArguments[const Symbol('data')] as Map);
        return Response(
          data: createdGroupResponse(),
          statusCode: 201,
          requestOptions: RequestOptions(path: '/groups'),
        );
      });

      final repo = GroupsRepository(ApiService(mockDio));
      await repo.createGroup(name: 'Группа', categories: ['FUNNY']);
      expect(body!.containsKey('kind_only'), isFalse);

      await repo.createGroup(
          name: 'Группа', categories: ['FUNNY'], kindOnly: true);
      expect(body!['kind_only'], isTrue);
    });

    test('setKindOnly patches the group and returns the stored value',
        () async {
      final mockDio = MockDio();
      when(() => mockDio.patch('/groups/g1', data: any(named: 'data')))
          .thenAnswer((_) async => Response(
                data: {
                  'data': {
                    'group': {
                      'id': 'g1',
                      'name': 'Группа',
                      'admin_id': 'u1',
                      'invite_code': 'AB2CD3',
                      'categories': ['FUNNY'],
                      'kind_only': true,
                      'created_at': '2026-01-01T00:00:00Z',
                    }
                  }
                },
                statusCode: 200,
                requestOptions: RequestOptions(path: '/groups/g1'),
              ));

      final repo = GroupsRepository(ApiService(mockDio));
      final group = await repo.setKindOnly('g1', true);
      expect(group.kindOnly, isTrue);
    });
  });

  group('the Group model', () {
    test('round-trips kind_only', () {
      const group = Group(
        id: 'g1',
        name: 'Группа',
        adminId: 'u1',
        inviteCode: 'AB2CD3',
        categories: ['FUNNY'],
        kindOnly: true,
        createdAt: '2026-01-01T00:00:00Z',
      );
      expect(Group.fromJson(group.toJson()).kindOnly, isTrue);
    });

    test('a response without the field reads as an ordinary group', () {
      // An older backend, or a cached response from before the field existed, must not crash.
      final group = Group.fromJson({
        'id': 'g1',
        'name': 'Группа',
        'admin_id': 'u1',
        'invite_code': 'AB2CD3',
        'categories': ['FUNNY'],
        'created_at': '2026-01-01T00:00:00Z',
      });
      expect(group.kindOnly, isFalse);
    });
  });
}
