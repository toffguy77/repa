import 'package:dio/dio.dart';
import '../../../core/api/api_client.dart';
import '../../../core/api/api_service.dart';
import '../domain/voting.dart';

class VotingRepository {
  final ApiService _api;

  VotingRepository(this._api);

  Future<VotingSession> getVotingSession(String seasonId) async {
    try {
      final response = await _api.getVotingSession(seasonId);
      final data = response['data'] as Map<String, dynamic>;
      return VotingSession.fromJson(data);
    } on DioException catch (e) {
      throw parseError(e);
    }
  }

  Future<VoteResultData> castVote({
    required String seasonId,
    required String questionId,
    required String targetId,
  }) async {
    try {
      final response = await _api.castVote(seasonId, {
        'question_id': questionId,
        'target_id': targetId,
      });
      final data = response['data'] as Map<String, dynamic>;
      return VoteResultData.fromJson(data);
    } on DioException catch (e) {
      throw parseError(e);
    }
  }

  Future<GroupVotingProgress> getVotingProgress(String seasonId) async {
    try {
      final response = await _api.getVotingProgress(seasonId);
      final data = response['data'] as Map<String, dynamic>;
      return GroupVotingProgress.fromJson(data);
    } on DioException catch (e) {
      throw parseError(e);
    }
  }
}
