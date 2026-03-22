import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:repa/core/api/api_service.dart';
import 'package:repa/features/groups/data/groups_repository.dart';
import 'package:repa/features/groups/presentation/groups_notifier.dart';
import 'package:repa/features/home/home_screen.dart';
import 'package:mocktail/mocktail.dart';

class _MockApiService extends Mock implements ApiService {}

class _EmptyGroupsListNotifier extends GroupsListNotifier {
  _EmptyGroupsListNotifier() : super(GroupsRepository(_MockApiService()));

  @override
  Future<void> load() async {
    state = const GroupsListState(groups: []);
  }

  @override
  Future<void> refresh() async {
    state = const GroupsListState(groups: []);
  }
}

void main() {
  Widget buildApp() {
    return ProviderScope(
      overrides: [
        groupsListProvider.overrideWith((ref) => _EmptyGroupsListNotifier()),
      ],
      child: const MaterialApp(home: HomeScreen()),
    );
  }

  testWidgets('renders groups tab with title and bottom nav', (tester) async {
    await tester.pumpWidget(buildApp());
    await tester.pump();

    expect(find.text('Мои группы'), findsWidgets);
    expect(find.byIcon(Icons.group), findsOneWidget);
    expect(find.byIcon(Icons.person), findsOneWidget);
  });

  testWidgets('shows empty state when no groups', (tester) async {
    await tester.pumpWidget(buildApp());
    await tester.pump();

    expect(find.text('Пока нет групп'), findsOneWidget);
    expect(find.text('Создать группу'), findsOneWidget);
  });

  testWidgets('profile tab shows placeholder and logout', (tester) async {
    await tester.pumpWidget(buildApp());
    await tester.pump();

    await tester.tap(find.text('Профиль'));
    await tester.pump();

    expect(find.text('Скоро здесь будет профиль'), findsOneWidget);
    expect(find.text('Выйти'), findsOneWidget);
  });
}
