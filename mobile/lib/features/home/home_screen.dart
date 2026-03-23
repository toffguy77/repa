import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../core/providers/auth_provider.dart';
import '../../core/providers/connectivity_provider.dart';
import '../../core/theme/app_colors.dart';
import '../../core/theme/app_text_styles.dart';
import '../../core/widgets/empty_state_widget.dart';
import '../../core/widgets/error_state_widget.dart';
import '../../core/widgets/skeleton_loader.dart';
import '../groups/presentation/groups_notifier.dart';
import '../crystals/presentation/widgets/crystal_balance_widget.dart';
import '../groups/presentation/widgets/group_card.dart';

class HomeScreen extends ConsumerStatefulWidget {
  const HomeScreen({super.key});

  @override
  ConsumerState<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends ConsumerState<HomeScreen> {
  int _tabIndex = 0;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(groupsListProvider.notifier).load();
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: IndexedStack(
        index: _tabIndex,
        children: const [
          _GroupsTab(),
          _ProfileTab(),
        ],
      ),
      floatingActionButton: _tabIndex == 0
          ? FloatingActionButton(
              onPressed: () => context.push('/groups/create'),
              backgroundColor: AppColors.primary,
              child: const Icon(Icons.add, color: Colors.white),
            )
          : null,
      bottomNavigationBar: BottomNavigationBar(
        currentIndex: _tabIndex,
        onTap: (i) => setState(() => _tabIndex = i),
        selectedItemColor: AppColors.primary,
        items: const [
          BottomNavigationBarItem(
            icon: Icon(Icons.group),
            label: 'Мои группы',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.person),
            label: 'Профиль',
          ),
        ],
      ),
    );
  }
}

class _GroupsTab extends ConsumerWidget {
  const _GroupsTab();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(groupsListProvider);

    // Auto-refresh on reconnect
    ref.listen<bool>(connectivityProvider, (prev, next) {
      if (prev == false && next == true) {
        ref.read(groupsListProvider.notifier).refresh();
      }
    });

    return Scaffold(
      appBar: AppBar(
        title: const Text('Мои группы'),
        actions: [
          const CrystalBalanceWidget(),
          const SizedBox(width: 8),
          IconButton(
            icon: const Icon(Icons.person_add),
            onPressed: () => context.push('/groups/join'),
            tooltip: 'Вступить по ссылке',
          ),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: () => ref.read(groupsListProvider.notifier).refresh(),
        child: _buildBody(context, ref, state),
      ),
    );
  }

  Widget _buildBody(BuildContext context, WidgetRef ref, GroupsListState state) {
    if (state.loading && state.groups.isEmpty) {
      return ListView(
        padding: const EdgeInsets.only(top: 8),
        children: List.generate(4, (_) => const GroupCardSkeleton()),
      );
    }

    if (state.error != null && state.groups.isEmpty) {
      return ListView(
        children: [
          const SizedBox(height: 120),
          ErrorStateWidget(
            message: state.error,
            onRetry: () => ref.read(groupsListProvider.notifier).load(),
          ),
        ],
      );
    }

    if (state.groups.isEmpty) {
      return ListView(
        children: [
          const SizedBox(height: 120),
          EmptyStateWidget(
            emoji: '\u{1F351}',
            title: 'Пока нет групп',
            subtitle: 'Создай группу или вступи по ссылке',
            buttonText: 'Создать группу',
            onButtonPressed: () => context.push('/groups/create'),
          ),
        ],
      );
    }

    return ListView.builder(
      padding: const EdgeInsets.only(top: 8, bottom: 80),
      itemCount: state.groups.length,
      itemBuilder: (context, index) {
        final group = state.groups[index];
        return GroupCard(
          group: group,
          onTap: () => context.push('/groups/${group.id}'),
        );
      },
    );
  }
}

// TODO: implement profile tab (T14)
class _ProfileTab extends ConsumerWidget {
  const _ProfileTab();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Scaffold(
      appBar: AppBar(title: const Text('Профиль')),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Text('\u{1F351}', style: TextStyle(fontSize: 64)),
            const SizedBox(height: 16),
            Text('Скоро здесь будет профиль', style: AppTextStyles.caption),
            const SizedBox(height: 32),
            OutlinedButton(
              onPressed: () => ref.read(authProvider.notifier).logout(),
              child: Text(
                'Выйти',
                style: TextStyle(color: AppColors.error),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
