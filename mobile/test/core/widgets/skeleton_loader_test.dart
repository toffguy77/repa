import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:repa/core/widgets/skeleton_loader.dart';
import 'package:repa/core/theme/app_theme.dart';

void main() {
  // SkeletonLoader uses flutter_animate's shimmer with repeat().
  // After assertions, we replace the widget and settle to drain pending timers.

  Future<void> cleanUp(WidgetTester tester) async {
    await tester.pumpWidget(MaterialApp(theme: AppTheme.dark, home: SizedBox()));
    await tester.pumpAndSettle();
  }

  testWidgets('GroupCardSkeleton renders', (tester) async {
    await tester.pumpWidget(
      MaterialApp(theme: AppTheme.dark, home: Scaffold(body: GroupCardSkeleton())),
    );
    expect(find.byType(GroupCardSkeleton), findsOneWidget);
    await cleanUp(tester);
  });

  testWidgets('MemberAvatarSkeleton renders', (tester) async {
    await tester.pumpWidget(
      MaterialApp(theme: AppTheme.dark, home: Scaffold(body: MemberAvatarSkeleton())),
    );
    expect(find.byType(MemberAvatarSkeleton), findsOneWidget);
    await cleanUp(tester);
  });

  testWidgets('MemberCardSkeleton renders', (tester) async {
    await tester.pumpWidget(
      MaterialApp(theme: AppTheme.dark, home: Scaffold(body: MemberCardSkeleton())),
    );
    expect(find.byType(MemberCardSkeleton), findsOneWidget);
    await cleanUp(tester);
  });

  testWidgets('VotingSessionSkeleton renders', (tester) async {
    await tester.pumpWidget(
      MaterialApp(theme: AppTheme.dark, home: Scaffold(body: VotingSessionSkeleton())),
    );
    expect(find.byType(VotingSessionSkeleton), findsOneWidget);
    await cleanUp(tester);
  });
}
