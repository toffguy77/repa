import '../../../core/api/api_service.dart';
import '../domain/question_candidate.dart';

class QuestionVoteRepository {
  final ApiService _api;

  QuestionVoteRepository(this._api);

  Future<List<QuestionCandidate>> getCandidates(String groupId) async {
    final result = await _api.getQuestionCandidates(groupId);
    final data = result['data'] as Map<String, dynamic>;
    final list = data['candidates'] as List<dynamic>;
    return list
        .map((e) => QuestionCandidate.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<void> vote(String groupId, String questionId) async {
    await _api.voteQuestion(groupId, questionId);
  }
}
