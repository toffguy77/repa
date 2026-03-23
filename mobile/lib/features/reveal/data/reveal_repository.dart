import '../../../core/api/api_service.dart';
import '../domain/reveal.dart';

class RevealRepository {
  final ApiService _api;

  RevealRepository(this._api);

  Future<RevealData> getReveal(String seasonId) async {
    final response = await _api.getReveal(seasonId);
    final data = response['data'] as Map<String, dynamic>;
    return RevealData.fromJson(data);
  }

  Future<List<MemberCard>> getMembersCards(String seasonId) async {
    final response = await _api.getMembersCards(seasonId);
    final data = response['data'] as Map<String, dynamic>;
    final list = data['members'] as List<dynamic>;
    return list
        .map((e) => MemberCard.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<RevealData> openHidden(String seasonId) async {
    final response = await _api.openHidden(seasonId);
    // Returns all_attributes + crystal_balance, but we re-fetch full reveal
    // to keep state consistent
    final _ = response;
    return getReveal(seasonId);
  }

  Future<CardUrlResult> getMyCardUrl(String seasonId) async {
    final response = await _api.getMyCardUrl(seasonId);
    final data = response['data'] as Map<String, dynamic>;
    return CardUrlResult.fromJson(data);
  }

  Future<DetectorResult> getDetector(String seasonId) async {
    final response = await _api.getDetector(seasonId);
    final data = response['data'] as Map<String, dynamic>;
    return DetectorResult.fromJson(data);
  }

  Future<DetectorResult> buyDetector(String seasonId) async {
    final response = await _api.buyDetector(seasonId);
    final data = response['data'] as Map<String, dynamic>;
    return DetectorResult.fromJson(data);
  }

  Future<ReactionCounts> getReactions(
      String seasonId, String targetId) async {
    final response = await _api.getReactions(seasonId, targetId);
    final data = response['data'] as Map<String, dynamic>;
    return ReactionCounts.fromJson(data);
  }

  Future<ReactionCounts> createReaction(
      String seasonId, String targetId, String emoji) async {
    final response = await _api.createReaction(seasonId, targetId, emoji);
    final data = response['data'] as Map<String, dynamic>;
    return ReactionCounts.fromJson(data);
  }
}
