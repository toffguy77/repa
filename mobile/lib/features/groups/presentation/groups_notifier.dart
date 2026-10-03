import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/api/api_client.dart';
import '../../../core/providers/api_provider.dart';
import '../data/groups_repository.dart';
import '../domain/group.dart';
import '../../reveal/domain/share_link.dart';

final groupsRepositoryProvider = Provider<GroupsRepository>((ref) {
  final api = ref.watch(apiServiceProvider);
  return GroupsRepository(api);
});

// --- Groups List ---

class GroupsListState {
  final bool loading;
  final String? error;
  final List<GroupListItem> groups;

  const GroupsListState({
    this.loading = false,
    this.error,
    this.groups = const [],
  });
}

class GroupsListNotifier extends StateNotifier<GroupsListState> {
  final GroupsRepository _repo;

  GroupsListNotifier(this._repo) : super(const GroupsListState());

  Future<void> load() async {
    state = GroupsListState(loading: true, groups: state.groups);
    try {
      final groups = await _repo.listGroups();
      state = GroupsListState(groups: groups);
    } on AppException catch (e) {
      state = GroupsListState(error: e.message, groups: state.groups);
    }
  }

  Future<void> refresh() async {
    try {
      final groups = await _repo.listGroups();
      state = GroupsListState(groups: groups);
    } on AppException catch (e) {
      state = GroupsListState(error: e.message, groups: state.groups);
    }
  }
}

final groupsListProvider =
    StateNotifierProvider<GroupsListNotifier, GroupsListState>((ref) {
  return GroupsListNotifier(ref.watch(groupsRepositoryProvider));
});

// --- Create Group ---

class CreateGroupState {
  final bool loading;
  final String? error;

  const CreateGroupState({this.loading = false, this.error});
}

class CreateGroupNotifier extends StateNotifier<CreateGroupState> {
  final GroupsRepository _repo;

  CreateGroupNotifier(this._repo) : super(const CreateGroupState());

  Future<CreateGroupResult?> create({
    required String name,
    required List<String> categories,
    String? telegramUsername,
    bool? kindOnly,
  }) async {
    state = const CreateGroupState(loading: true);
    try {
      final result = await _repo.createGroup(
        name: name,
        categories: categories,
        telegramUsername: telegramUsername,
        kindOnly: kindOnly,
      );
      state = const CreateGroupState();
      return result;
    } on AppException catch (e) {
      state = CreateGroupState(error: e.message);
      return null;
    }
  }
}

final createGroupProvider =
    StateNotifierProvider<CreateGroupNotifier, CreateGroupState>((ref) {
  return CreateGroupNotifier(ref.watch(groupsRepositoryProvider));
});

// --- Join Group ---

class JoinGroupState {
  final bool loading;
  final bool previewing;
  final String? error;
  final JoinPreview? preview;

  const JoinGroupState({
    this.loading = false,
    this.previewing = false,
    this.error,
    this.preview,
  });
}

class JoinGroupNotifier extends StateNotifier<JoinGroupState> {
  final GroupsRepository _repo;

  JoinGroupNotifier(this._repo) : super(const JoinGroupState());

  void reset() {
    state = const JoinGroupState();
  }

  String _extractCode(String input) {
    final uri = Uri.tryParse(input);
    if (uri != null && uri.pathSegments.length >= 2 && uri.pathSegments[0] == 'join') {
      return uri.pathSegments[1];
    }
    if (input.contains('/join/')) {
      return input.split('/join/').last.trim();
    }
    return input.trim();
  }

  Future<void> loadPreview(String input) async {
    final code = _extractCode(input);
    if (code.isEmpty) {
      state = const JoinGroupState();
      return;
    }
    state = JoinGroupState(previewing: true);
    try {
      final preview = await _repo.joinPreview(code);
      state = JoinGroupState(preview: preview);
    } on AppException catch (e) {
      state = JoinGroupState(error: e.message);
    }
  }

  /// [source] attributes the join. Defaults to the typed-code path, which is what the join
  /// screen is; a deep link passes the source its marker implies.
  Future<Group?> join(
    String input, {
    JoinSource source = JoinSource.code,
    String? referrerId,
  }) async {
    final code = _extractCode(input);
    state = JoinGroupState(loading: true, preview: state.preview);
    try {
      final group = await _repo.joinGroup(
        code,
        source: source.wireValue,
        referrerId: referrerId,
      );
      state = const JoinGroupState();
      return group;
    } on AppException catch (e) {
      state = JoinGroupState(error: e.message, preview: state.preview);
      return null;
    }
  }
}

final joinGroupProvider =
    StateNotifierProvider<JoinGroupNotifier, JoinGroupState>((ref) {
  return JoinGroupNotifier(ref.watch(groupsRepositoryProvider));
});

// --- Group Detail ---

class GroupDetailState {
  final bool loading;
  final String? error;
  final GroupDetail? detail;

  const GroupDetailState({this.loading = false, this.error, this.detail});
}

class GroupDetailNotifier extends StateNotifier<GroupDetailState> {
  final GroupsRepository _repo;
  final String groupId;

  GroupDetailNotifier(this._repo, this.groupId)
      : super(const GroupDetailState());

  Future<void> load() async {
    state = GroupDetailState(loading: true, detail: state.detail);
    try {
      final detail = await _repo.getGroup(groupId);
      state = GroupDetailState(detail: detail);
    } on AppException catch (e) {
      state = GroupDetailState(error: e.message, detail: state.detail);
    }
  }

  /// [permanent] records that the departure is irreversible — an invite will not bring the member
  /// back. Ordinary leaving stays reversible on purpose.
  Future<bool> leave({bool permanent = false}) async {
    try {
      await _repo.leaveGroup(groupId, permanent: permanent);
      return true;
    } on AppException {
      return false;
    }
  }

  Future<String?> regenerateInvite() async {
    try {
      return await _repo.regenerateInviteLink(groupId);
    } on AppException {
      return null;
    }
  }

  /// Flips the kind-only setting. Returns the server's message on refusal — the one case the member
  /// can act on is a category set that leaves nothing kind to ask, and they can only fix it if told.
  Future<String?> setKindOnly(bool kindOnly) async {
    try {
      await _repo.setKindOnly(groupId, kindOnly);
      await load();
      return null;
    } on AppException catch (e) {
      return e.message;
    }
  }
}

final groupDetailProvider = StateNotifierProvider.autoDispose
    .family<GroupDetailNotifier, GroupDetailState, String>((ref, groupId) {
  return GroupDetailNotifier(ref.watch(groupsRepositoryProvider), groupId);
});
