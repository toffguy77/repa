import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/api/api_client.dart';
import '../../../core/providers/api_provider.dart';
import '../data/profile_repository.dart';
import '../domain/profile.dart';

final profileRepositoryProvider = Provider<ProfileRepository>((ref) {
  final api = ref.watch(apiServiceProvider);
  return ProfileRepository(api);
});

class ProfileState {
  final bool loading;
  final String? error;
  final MemberProfile? profile;

  const ProfileState({this.loading = false, this.error, this.profile});
}

class ProfileNotifier extends StateNotifier<ProfileState> {
  final ProfileRepository _repo;
  final String groupId;
  final String userId;

  ProfileNotifier(this._repo, this.groupId, this.userId)
      : super(const ProfileState());

  Future<void> load() async {
    state = ProfileState(loading: true, profile: state.profile);
    try {
      final profile = await _repo.getMemberProfile(groupId, userId);
      state = ProfileState(profile: profile);
    } on AppException catch (e) {
      state = ProfileState(error: e.message, profile: state.profile);
    } catch (e) {
      state = ProfileState(
          error: 'Что-то пошло не так', profile: state.profile);
    }
  }
}

final profileProvider = StateNotifierProvider.autoDispose
    .family<ProfileNotifier, ProfileState, ({String groupId, String userId})>(
        (ref, args) {
  return ProfileNotifier(
    ref.watch(profileRepositoryProvider),
    args.groupId,
    args.userId,
  );
});
